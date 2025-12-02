package coco

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
	default:
		return ""
	}
}

const (
	SideLeft   Side = "left"
	SideRight  Side = "right"
	SideTop    Side = "top"
	SideBottom Side = "bottom"
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
