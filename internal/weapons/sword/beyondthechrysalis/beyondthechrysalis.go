package beyondthechrysalis

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	devotionKey  = "beyondthechrysalis-devotion"
	defianceKey  = "beyondthechrysalis-defiance"
	buffDuration = 10 * 60
)

func init() {
	core.RegisterWeaponFunc(keys.BeyondTheChrysalis, NewWeapon)
}

type Weapon struct {
	Index    int
	core     *core.Core
	char     *character.CharWrapper
	refine   int
	sequence int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		char:   char,
		core:   c,
		refine: p.Refine,
	}

	cdBuff := 0.40 + float64(w.refine-1)*0.16
	swirlBuff := 0.27 + float64(w.refine-1)*0.09
	maxEnergy := 4.5 + float64(w.refine-1)*0.5

	m := make([]float64, attributes.EndStatType)
	dmgMod := make([]float64, attributes.EndStatType)

	// 1. Stat Modifier for CRIT DMG (Winds of Devotion)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("beyondthechrysalis-stats", -1),
		AffectedStat: attributes.CD,
		Amount: func() []float64 {
			m[attributes.CD] = 0
			if char.StatusIsActive(devotionKey) {
				m[attributes.CD] = cdBuff
			}
			return m
		},
	})

	// 2. Attack Modifier for Stellar Swirl DMG (Winds of Defiance)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("beyondthechrysalis-defiance", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			dmgMod[attributes.DmgP] = 0
			
			// If buff is active AND the attack is tagged as Stellar Swirl
			if char.StatusIsActive(defianceKey) && atk.Info.AttackTag == attacks.AttackTagDirectStellarSwirl {
				dmgMod[attributes.DmgP] = swirlBuff
				return dmgMod, true
			}
			return nil, false
		},
	})

	onSkillOrBurst := func(args ...any) {
		if c.Player.Active() != char.Index() {
			return
		}
		switch w.sequence % 3 {
		case 0:
			char.AddStatus(devotionKey, buffDuration, true)
		case 1:
			char.AddStatus(defianceKey, buffDuration, true)
		case 2:
			char.AddEnergy("beyondthechrysalis-plenty", maxEnergy)
		}
		w.sequence++
	}

	c.Events.Subscribe(event.OnSkill, onSkillOrBurst, fmt.Sprintf("beyondthechrysalis-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnBurst, onSkillOrBurst, fmt.Sprintf("beyondthechrysalis-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		prev := args[0].(int)
		if prev == char.Index() {
			w.sequence = 0
		}
	}, fmt.Sprintf("beyondthechrysalis-swap-%v", char.Base.Key.String()))

	return w, nil
}
