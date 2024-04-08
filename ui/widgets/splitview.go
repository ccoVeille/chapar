package widgets

import (
	"image"

	"github.com/mirzakhany/chapar/ui/chapartheme"

	"gioui.org/op/paint"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
)

type SplitView struct {
	// Ratio keeps the current layout.
	// 0 is center, -1 completely to the left, 1 completely to the right.
	Ratio float32

	drag   bool
	dragID pointer.ID
	dragX  float32

	// Bar is the width for resizing the layout
	BarWidth unit.Dp

	MinLeftSize unit.Dp
	MaxLeftSize unit.Dp

	MinRightSize unit.Dp
	MaxRightSize unit.Dp
}

const defaultBarWidth = unit.Dp(2)

func (s *SplitView) Layout(gtx layout.Context, theme *chapartheme.Theme, left, right layout.Widget) layout.Dimensions {
	bar := gtx.Dp(s.BarWidth)
	if bar <= 1 {
		bar = gtx.Dp(defaultBarWidth)
	}

	mils := gtx.Dp(s.MinLeftSize)
	mals := gtx.Dp(s.MaxLeftSize)
	mirs := gtx.Dp(s.MinRightSize)
	mars := gtx.Dp(s.MaxRightSize)

	// 0.18 := (x + 1) / 2
	proportion := (s.Ratio + 1) / 2
	leftSize := int(proportion*float32(gtx.Constraints.Max.X) - float32(bar))
	if leftSize < mils {
		leftSize = mils
	}

	if leftSize > mals && mals > 0 {
		leftSize = mals
	}

	rightOffset := leftSize + bar
	rightsize := gtx.Constraints.Max.X - rightOffset

	if rightsize < mirs {
		rightsize = mirs
	}

	if rightsize > mars && mars > 0 {
		rightsize = mars
	}

	{
		barColor := theme.SeparatorColor
		// register for input
		barRect := image.Rect(leftSize, 0, rightOffset, gtx.Constraints.Max.X)
		area := clip.Rect(barRect).Push(gtx.Ops)
		// paint.FillShape(gtx.Ops, barColor, clip.Rect(barRect).Op())
		// defer clip.UniformRRect(barRect, 0).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, theme.SideBarBgColor)

		pointer.CursorColResize.Add(gtx.Ops)
		event.Op(gtx.Ops, s)

		for {
			ev, ok := gtx.Event(
				pointer.Filter{
					Target: s,
					Kinds:  pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel,
				},
			)

			if !ok {
				break
			}

			e, ok := ev.(pointer.Event)
			if !ok {
				continue
			}

			switch e.Kind {
			case pointer.Press:
				if s.drag {
					break
				}

				barColor = Hovered(barColor)
				s.dragID = e.PointerID
				s.dragX = e.Position.X

			case pointer.Drag:
				if s.dragID != e.PointerID {
					break
				}

				// if barColor != s.BarColorHover {
				barColor = Hovered(barColor)
				// }

				deltaX := e.Position.X - s.dragX
				s.dragX = e.Position.X

				deltaRatio := deltaX * 2 / float32(gtx.Constraints.Max.X)
				s.Ratio += deltaRatio

				if e.Priority < pointer.Grabbed {
					gtx.Execute(pointer.GrabCmd{
						Tag: s,
						ID:  s.dragID,
					})
				}

			case pointer.Release:
				barColor = theme.SeparatorColor
				fallthrough
			case pointer.Cancel:
				s.drag = false
				barColor = theme.SeparatorColor
			default:

				continue
			}
		}
		paint.FillShape(gtx.Ops, barColor, clip.Rect(barRect).Op())
		area.Pop()
	}

	{
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(leftSize, gtx.Constraints.Max.Y))
		left(gtx)
	}

	{
		off := op.Offset(image.Pt(rightOffset, 0)).Push(gtx.Ops)
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(rightsize, gtx.Constraints.Max.Y))
		right(gtx)
		off.Pop()
	}

	return layout.Dimensions{Size: gtx.Constraints.Max}
}
