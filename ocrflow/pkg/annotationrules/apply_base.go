package annotationrules

type ContactType string

const (
	ContactTypeInner ContactType = "inner"
	ContactTypeOuter ContactType = "outer"
)

type Side string

func SideFromString(s string) Side {
	switch s {
	case "left":
		return SideLeft
	case "right":
		return SideRight
	case "top":
		return SideTop
	case "bottom":
		return SideBottom
	case "horizontal":
		return SideHorizontal
	case "vertical":
		return SideVertical
	case "all":
		return SideAll
	default:
		return ""
	}
}

const (
	SideLeft       Side = "left"
	SideRight      Side = "right"
	SideTop        Side = "top"
	SideBottom     Side = "bottom"
	SideHorizontal Side = "horizontal"
	SideVertical   Side = "vertical"
	SideAll        Side = "all"
)

func ContactTypeFromString(s string) ContactType {
	switch s {
	case "inner":
		return ContactTypeInner
	case "outer":
		return ContactTypeOuter
	default:
		return ""
	}
}
