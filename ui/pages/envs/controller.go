package envs

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mirzakhany/chapar/internal/domain"
	"github.com/mirzakhany/chapar/internal/loader"
)

type Controller struct {
	view  *View
	model *Model

	activeTabID string
}

func NewController(view *View, model *Model) *Controller {
	c := &Controller{
		view:  view,
		model: model,
	}

	model.AddEnvChangeListener(c.onEnvChange)
	view.SetOnNewEnv(c.onNewEnv)
	view.SetOnTitleChanged(c.onTitleChanged)
	view.SetOnTreeViewNodeDoubleClicked(c.onTreeViewNodeDoubleClicked)
	view.SetOnTabSelected(c.onTabSelected)
	view.SetOnItemsChanged(c.onItemsChanged)
	view.SetOnSave(c.onSave)
	view.SetOnTabClose(c.onTabClose)
	view.SetOnTreeViewMenuClicked(c.onTreeViewMenuClicked)
	return c
}

func (c *Controller) onEnvChange(env *domain.Environment) {
	c.model.UpdateEnvironment(env)
}

func (c *Controller) onNewEnv() {
	env := domain.NewEnvironment("New Environment")
	c.model.AddEnvironment(env)
}

func (c *Controller) onTitleChanged(id string, title string) {
	env := c.model.GetEnvironment(id)
	if env == nil {
		return
	}
	env.MetaData.Name = title
	c.view.UpdateTreeNodeTitle(env.MetaData.ID, env.MetaData.Name)
	c.view.UpdateTabTitle(env.MetaData.ID, env.MetaData.Name)
	c.model.UpdateEnvironment(env)
	c.saveEnvironmentToDisc(id)
}

func (c *Controller) onTreeViewNodeDoubleClicked(id string) {
	env := c.model.GetEnvironment(id)
	if env == nil {
		return
	}

	if c.view.IsEnvTabOpen(id) {
		c.view.SwitchToTab(env)
		c.view.OpenContainer(env)
		return
	}

	c.view.OpenTab(env)
	c.view.OpenContainer(env)
}

func (c *Controller) LoadData() error {
	data, err := loader.ReadEnvironmentsData()
	if err != nil {
		return err
	}
	c.model.SetEnvironments(data)
	c.view.PopulateTreeView(data)
	return nil
}

func (c *Controller) onTabSelected(id string) {
	if c.activeTabID == id {
		return
	}
	c.activeTabID = id
	env := c.model.GetEnvironment(id)
	c.view.SwitchToTab(env)
	c.view.OpenContainer(env)
}

func (c *Controller) onItemsChanged(id string, items []domain.KeyValue) {
	env := c.model.GetEnvironment(id)
	if env == nil {
		return
	}

	// is data changed?
	if domain.CompareKeyValues(env.Spec.Values, items) {
		return
	}

	env.Spec.Values = items
	c.model.UpdateEnvironment(env)

	// set tab dirty if the in memory data is different from the file
	envFromFile, err := c.model.GetEnvironmentFromDisc(id)
	if err != nil {
		fmt.Println("failed to get environment from file", err)
		return
	}

	c.view.SetTabDirty(id, !domain.CompareKeyValues(env.Spec.Values, envFromFile.Spec.Values))
}

func (c *Controller) onSave(id string) {
	c.saveEnvironmentToDisc(id)
}

func (c *Controller) onTabClose(id string) {
	c.view.CloseTab(id)
}

func (c *Controller) saveEnvironmentToDisc(id string) {
	env := c.model.GetEnvironment(id)
	if env == nil {
		return
	}
	if err := loader.UpdateEnvironment(env); err != nil {
		fmt.Println("failed to update environment", err)
		return
	}
	c.view.SetTabDirty(id, false)
}

func (c *Controller) onTreeViewMenuClicked(id string, action string) {
	switch action {
	case Duplicate:
		c.duplicateEnvironment(id)
	case Delete:
		c.deleteEnvironment(id)
	}
}

func (c *Controller) duplicateEnvironment(id string) {
	// read environment from file to make sure we have the latest persisted data
	envFromFile, err := c.model.GetEnvironmentFromDisc(id)
	if err != nil {
		fmt.Println("failed to get environment from file", err)
		return
	}

	newEnv := envFromFile.Clone()
	newEnv.MetaData.ID = uuid.NewString()
	newEnv.MetaData.Name = newEnv.MetaData.Name + " (copy)"
	newEnv.FilePath = loader.AddSuffixBeforeExt(newEnv.FilePath, "-copy")
	c.model.AddEnvironment(newEnv)
	c.view.AddTreeViewNode(newEnv)
	c.saveEnvironmentToDisc(newEnv.MetaData.ID)
}

func (c *Controller) deleteEnvironment(id string) {
	env := c.model.GetEnvironment(id)
	if env == nil {
		return
	}
	c.model.DeleteEnvironment(id)
	c.view.RemoveTreeViewNode(id)
	if err := loader.DeleteEnvironment(env); err != nil {
		fmt.Println("failed to delete environment", err)
	}
	c.view.CloseTab(id)
}
