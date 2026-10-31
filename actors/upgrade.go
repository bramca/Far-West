package actors

type Upgrade int

const (
	UpgradeMaxHealth Upgrade = iota
	UpgradeDamage
	UpgradeReload
	UpgradeSpeed
	UpgradeDodge
	UpgradeArmor
)

var AllUpgrades = []Upgrade{
	UpgradeMaxHealth,
	UpgradeDamage,
	UpgradeReload,
	UpgradeSpeed,
	UpgradeDodge,
	UpgradeArmor,
}

func (u Upgrade) String() string {
	switch u {
	case UpgradeMaxHealth:
		return "IRON CONSTITUTION"
	case UpgradeDamage:
		return "DEAD EYE"
	case UpgradeReload:
		return "QUICK HANDS"
	case UpgradeSpeed:
		return "LIGHT FEET"
	case UpgradeDodge:
		return "COOL HEAD"
	case UpgradeArmor:
		return "TOUGH HIDE"
	default:
		return "UNKNOWN UPGRADE"
	}
}

func (u Upgrade) Description() string {
	switch u {
	case UpgradeMaxHealth:
		return "+5 MAX HEALTH"
	case UpgradeDamage:
		return "+1 BULLET DAMAGE"
	case UpgradeReload:
		return "-10 RELOAD FRAMES"
	case UpgradeSpeed:
		return "+0.25 MOVE SPEED"
	case UpgradeDodge:
		return "-6 DODGE COOLDOWN"
	case UpgradeArmor:
		return "-1 DAMAGE TAKEN"
	default:
		return ""
	}
}

func (u Upgrade) BasePrice() int {
	switch u {
	case UpgradeMaxHealth:
		return 220
	case UpgradeDamage:
		return 260
	case UpgradeReload:
		return 180
	case UpgradeSpeed:
		return 180
	case UpgradeDodge:
		return 180
	case UpgradeArmor:
		return 300
	default:
		return 200
	}
}
