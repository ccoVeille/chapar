package requests

import (
	"fmt"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/google/uuid"
	"github.com/mirzakhany/chapar/internal/domain"
	"github.com/mirzakhany/chapar/internal/loader"
	"github.com/mirzakhany/chapar/internal/logger"
	"github.com/mirzakhany/chapar/ui/pages/requests/rest"
	"github.com/mirzakhany/chapar/ui/widgets"
)

type Requests struct {
	theme *material.Theme

	addRequestButton widget.Clickable
	importButton     widget.Clickable
	searchBox        *widgets.TextField

	treeView *widgets.TreeView

	split widgets.SplitView
	tabs  *widgets.Tabs

	// collections represents the collections of requests
	collections []*domain.Collection
	// requests represents the standalone requests that are not in any collection
	requests []*domain.Request

	openedTabs []*openedTab

	selectedIndex int
}

type openedTab struct {
	req       *domain.Request
	tab       *widgets.Tab
	listItem  *widgets.TreeNode
	container Container

	closed bool
}

func New(theme *material.Theme) (*Requests, error) {
	collections, err := loader.LoadCollections()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to load collections, err %v", err))
		return nil, err
	}

	requests, err := loader.LoadRequests()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to load requests, err %v", err))
		return nil, err
	}

	logger.Info("collections and requests are loaded")

	search := widgets.NewTextField("", "Search...")
	search.SetIcon(widgets.SearchIcon, widgets.IconPositionEnd)
	search.SetBorderColor(widgets.Gray600)

	req := &Requests{
		theme:       theme,
		collections: collections,
		requests:    requests,
		searchBox:   search,
		tabs:        widgets.NewTabs([]*widgets.Tab{}, nil),
		treeView:    widgets.NewTreeView(prepareTreeView(collections, requests)),
		split: widgets.SplitView{
			Ratio:         -0.64,
			MinLeftSize:   unit.Dp(250),
			MaxLeftSize:   unit.Dp(800),
			BarWidth:      unit.Dp(2),
			BarColor:      color.NRGBA{R: 0x2b, G: 0x2d, B: 0x31, A: 0xff},
			BarColorHover: theme.Palette.ContrastBg,
		},
		openedTabs: make([]*openedTab, 0),
	}
	req.treeView.ParentMenuOptions = []string{"Duplicate", "Rename", "Delete"}
	req.treeView.ChildMenuOptions = []string{"Move", "Duplicate", "Rename", "Delete"}
	req.treeView.OnDoubleClick(req.onItemDoubleClick)
	req.treeView.SetOnMenuItemClick(func(tr *widgets.TreeNode, item string) {
		if item == "Duplicate" {
			req.duplicateReq(tr.Identifier)
		}

		if item == "Delete" {
			req.deleteReq(tr.Identifier)
		}
	})
	req.searchBox.SetOnTextChange(func(text string) {
		if req.collections == nil && req.requests == nil {
			return
		}
		req.treeView.Filter(text)
	})

	return req, nil
}

func prepareTreeView(collections []*domain.Collection, requests []*domain.Request) []*widgets.TreeNode {
	treeViewNodes := make([]*widgets.TreeNode, 0)
	for _, collection := range collections {
		parentNode := &widgets.TreeNode{
			Text:       collection.MetaData.Name,
			Identifier: collection.MetaData.ID,
			Children:   make([]*widgets.TreeNode, 0),
		}

		for _, req := range collection.Spec.Requests {
			if req.MetaData.ID == "" {
				req.MetaData.ID = uuid.NewString()
			}

			node := &widgets.TreeNode{
				Text:       req.MetaData.Name,
				Identifier: req.MetaData.ID,
			}

			parentNode.Children = append(parentNode.Children, node)
		}

		treeViewNodes = append(treeViewNodes, parentNode)
	}

	for _, req := range requests {
		node := &widgets.TreeNode{
			Text:       req.MetaData.Name,
			Identifier: req.MetaData.ID,
		}

		treeViewNodes = append(treeViewNodes, node)
	}

	return treeViewNodes
}

func (r *Requests) findRequestByID(id string) (*domain.Request, int) {
	for i, collection := range r.collections {
		for _, req := range collection.Spec.Requests {
			if req.MetaData.ID == id {
				return req, i
			}
		}
	}

	for _, req := range r.requests {
		if req.MetaData.ID == id {
			// -1 means this is a standalone request
			return req, -1
		}
	}

	return nil, -1
}

func (r *Requests) findRequestInTab(id string) (*openedTab, int) {
	for i, ot := range r.openedTabs {
		if ot.req.MetaData.ID == id {
			return ot, i
		}
	}
	return nil, -1
}

func (r *Requests) onTabClose(t *widgets.Tab) {
	tab, _ := r.findRequestInTab(t.Identifier)
	if tab != nil {
		if !tab.container.OnClose() {
			return
		}
		tab.closed = true
	}
}

func (r *Requests) onTitleChanged(id, title string) {
	// find the opened tab and mark it as dirty
	tab, _ := r.findRequestInTab(id)
	if tab != nil {
		if tab.req.MetaData.Name != title {
			// Update the tree view item and the tab title
			tab.req.MetaData.Name = title
			tab.tab.Title = title
			tab.listItem.Text = title
		}
	}
}

func (r *Requests) onItemDoubleClick(tr *widgets.TreeNode) {
	// if request is already opened, just switch to it
	tab, index := r.findRequestInTab(tr.Identifier)
	if tab != nil {
		r.selectedIndex = index
		r.tabs.SetSelected(index)
		return
	}

	req, _ := r.findRequestByID(tr.Identifier)
	if req != nil {
		tab := &widgets.Tab{Title: req.MetaData.Name, Closable: true, CloseClickable: &widget.Clickable{}}
		tab.SetOnClose(r.onTabClose)
		tab.SetIdentifier(req.MetaData.ID)

		ot := &openedTab{
			req:       req,
			tab:       tab,
			listItem:  tr,
			container: rest.NewRestContainer(r.theme, req.Clone()),
		}
		ot.container.SetOnTitleChanged(r.onTitleChanged)
		r.openedTabs = append(r.openedTabs, ot)

		i := r.tabs.AddTab(tab)
		r.selectedIndex = i
		r.tabs.SetSelected(i)
	}
}

func (r *Requests) duplicateReq(identifier string) {
	req, i := r.findRequestByID(identifier)
	if req != nil {
		newReq := req.Clone()
		newReq.MetaData.ID = uuid.NewString()
		newReq.MetaData.Name = newReq.MetaData.Name + " (copy)"
		// add copy to file name
		newReq.FilePath = loader.AddSuffixBeforeExt(newReq.FilePath, "-copy")

		node := &widgets.TreeNode{
			Text:       newReq.MetaData.Name,
			Identifier: newReq.MetaData.ID,
		}
		if i == -1 {
			r.requests = append(r.requests, newReq)
			r.treeView.AddNode(node)
		} else {
			r.collections[i].Spec.Requests = append(r.collections[i].Spec.Requests, newReq)
			r.treeView.AddChildNode(r.collections[i].MetaData.ID, node)
		}

		if err := loader.UpdateRequest(newReq); err != nil {
			logger.Error(fmt.Sprintf("failed to update request, err %v", err))
		}
	}
}

func (r *Requests) deleteReq(identifier string) {
	req, i := r.findRequestByID(identifier)
	if req != nil {
		if err := loader.DeleteRequest(req); err != nil {
			logger.Error(fmt.Sprintf("failed to delete request, err %v", err))
			return
		}

		if i == -1 {
			// TODO make it a function
			for j, item := range r.requests {
				if item.MetaData.ID == req.MetaData.ID {
					r.requests = append(r.requests[:j], r.requests[j+1:]...)
					break
				}
			}
		} else {
			r.collections[i].RemoveRequest(req)
		}
		r.treeView.RemoveNode(identifier)
	}
}

func (r *Requests) addNewEmptyReq(collectionID string) {
	req := domain.NewRequest("New Request")
	node := &widgets.TreeNode{
		Text:       req.MetaData.Name,
		Identifier: req.MetaData.ID,
	}

	if collectionID == "" {
		r.treeView.AddNode(node)
	} else {
		r.treeView.AddChildNode(collectionID, node)
	}

	tab := &widgets.Tab{Title: req.MetaData.Name, Closable: true, CloseClickable: &widget.Clickable{}}
	tab.SetOnClose(r.onTabClose)
	tab.SetIdentifier(req.MetaData.ID)

	ot := &openedTab{
		req:       req,
		tab:       tab,
		listItem:  node,
		container: rest.NewRestContainer(r.theme, req.Clone()),
	}
	ot.container.SetOnTitleChanged(r.onTitleChanged)
	r.openedTabs = append(r.openedTabs, ot)

	i := r.tabs.AddTab(tab)
	r.selectedIndex = i
	r.tabs.SetSelected(i)
}

func (r *Requests) list(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	return layout.Inset{Top: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(10), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Spacing: layout.SpaceStart}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if r.addRequestButton.Clicked(gtx) {
								r.addNewEmptyReq("")
							}

							return material.Button(theme, &r.addRequestButton, "Add").Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(2)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return material.Button(theme, &r.importButton, "Import").Layout(gtx)
						}),
					)
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(10), Left: unit.Dp(10), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return r.searchBox.Layout(gtx, theme)
				})
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(10), Right: 0}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return r.treeView.Layout(gtx, theme)
				})
			}),
		)
	})
}

func (r *Requests) container(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return r.tabs.Layout(gtx, theme)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if r.selectedIndex > len(r.openedTabs)-1 {
				return layout.Dimensions{}
			}
			ct := r.openedTabs[r.selectedIndex].container
			r.openedTabs[r.selectedIndex].tab.SetDirty(ct.IsDataChanged())

			return ct.Layout(gtx, theme)
		}),
	)
}

func (r *Requests) Layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	// update tabs with new items
	tabItems := make([]*widgets.Tab, 0)
	openItems := make([]*openedTab, 0)
	for _, ot := range r.openedTabs {
		if !ot.closed {
			tabItems = append(tabItems, ot.tab)
			openItems = append(openItems, ot)
		}
	}

	r.tabs.SetTabs(tabItems)
	r.openedTabs = openItems
	selectTab := r.tabs.Selected()
	gtx.Execute(op.InvalidateCmd{})

	// is selected tab is closed:
	// if its the last tab and there is another tab before it, select the previous one
	// if its the first tab and there is another tab after it, select the next one
	// if its the only tab, select it

	if selectTab > len(openItems)-1 {
		if len(openItems) > 0 {
			r.tabs.SetSelected(len(openItems) - 1)
		} else {
			selectTab = 0
			r.tabs.SetSelected(0)
		}
	}

	if r.selectedIndex != selectTab {
		r.selectedIndex = selectTab
	}

	return r.split.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			return r.list(gtx, theme)
		},
		func(gtx layout.Context) layout.Dimensions {
			return r.container(gtx, theme)
		},
	)
}
