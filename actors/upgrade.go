package actors

const (
	// DefaultDeadEyeMaxCharge is the size of the dead eye gauge a fresh
	// player starts with.
	DefaultDeadEyeMaxCharge = 100.0
	// DeadEyeChargeUpgrade is how much max charge the dead eye upgrade adds.
	DeadEyeChargeUpgrade = 25.0
)

type Upgrade int

const (
	UpgradeMaxHealth Upgrade = iota
	UpgradeDamage
	UpgradeReload
	UpgradeSpeed
	UpgradeDodge
	UpgradeArmor
	UpgradeFists
	UpgradeDeadEye
)

var AllUpgrades = []Upgrade{
	UpgradeMaxHealth,
	UpgradeDamage,
	UpgradeReload,
	UpgradeSpeed,
	UpgradeDodge,
	UpgradeArmor,
	UpgradeFists,
	UpgradeDeadEye,
}

func (u Upgrade) String() string {
	switch u {
	case UpgradeMaxHealth:
		return "IRON CONSTITUTION"
	case UpgradeDamage:
		return "GUNSLINGER"
	case UpgradeDeadEye:
		return "DEAD EYE"
	case UpgradeReload:
		return "QUICK HANDS"
	case UpgradeSpeed:
		return "LIGHT FEET"
	case UpgradeDodge:
		return "COOL HEAD"
	case UpgradeArmor:
		return "TOUGH HIDE"
	case UpgradeFists:
		return "IRON KNUCKLES"
	default:
		return "UNKNOWN UPGRADE"
	}
}

func (u Upgrade) Description() string {
	switch u {
	case UpgradeMaxHealth:
		return "+10 MAX HEALTH"
	case UpgradeDamage:
		return "+2 BULLET DAMAGE"
	case UpgradeReload:
		return "-20 RELOAD FRAMES"
	case UpgradeSpeed:
		return "+0.4 MOVE SPEED"
	case UpgradeDodge:
		return "-10 DODGE COOLDOWN"
	case UpgradeArmor:
		return "-2 DAMAGE TAKEN"
	case UpgradeFists:
		return "+4 FIST DAMAGE"
	case UpgradeDeadEye:
		return "+25 DEAD EYE CHARGE"
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
	case UpgradeFists:
		return 240
	case UpgradeDeadEye:
		return 260
	default:
		return 200
	}
}
