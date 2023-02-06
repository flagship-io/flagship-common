package targeting

type Strategy interface {
	ShouldIgnoreTargeting() bool
	Match() (bool, error)
}
