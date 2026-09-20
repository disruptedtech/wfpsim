package beyondthechrysalis

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	devotionKey    = "beyondthechrysalis-devotion"
	defianceKey    = "beyondthechrysalis-defiance"
	plentyKey      = "beyondthechrysalis-plenty"
	buffDuration   = 10 * 60
	energyInterval = 4 * 60
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

	// CD buff per refine: 56/72/88/104/120%
	cdBuff := 0.40 + float64(w.refine)*0.16

	// Stellar Swirl DMG buff per refine: 36/45/54/63/72%
	swirlBuff := 0.27 + float64(w.refine)*0.09

	// Max energy per refine: 5/5.5/6/6.5/7
	maxEnergy := 4.5 + float64(w.refine)*0.5

	c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		return w.onSkillOrBurst(cdBuff, swirlBuff, maxEnergy)
	}, "beyondthechrysalis-skill")

	c.Events.Subscribe(event.OnBurst, func(args ...interface{}) bool {
		return w.onSkillOrBurst(cdBuff, swirlBuff, maxEnergy)
	}, "beyondthechrysalis-burst")

	c.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		prev := args[0].(int)
		if prev == w.char.Index {
			w.sequence = 0
		}
		return false
	}, "beyondthechrysalis-swap")

	return w, nil
}

func (w *Weapon) onSkillOrBurst(cdBuff, swirlBuff, maxEnergy float64) bool {
	if w.core.Player.Active() != w.char.Index {
		return false
	}

	switch w.sequence % 3 {
	case 0: // Winds of Devotion - CRIT DMG buff
		val := make([]float64, attributes.EndStatType)
		val[attributes.CD] = cdBuff
		w.char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(devotionKey, buffDuration),
			AffectedStat: attributes.CD,
			Amount: func() ([]float64, bool) {
				return val, true
			},
		})

	case 1: // Winds of Defiance - Swirl DMG buff
		val := make([]float64, attributes.EndStatType)
		val[attributes.ElectroP] = swirlBuff
		w.char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(defianceKey, buffDuration),
			AffectedStat: attributes.ElectroP,
			Amount: func() ([]float64, bool) {
				return val, true
			},
		})

	case 2: // Winds of Plenty - Energy regen
		energyGiven := 0.0
		w.char.AddStatus(plentyKey, buffDuration, true)
		w.core.Events.Subscribe(event.OnTick, func(args ...interface{}) bool {
			if !w.char.StatusIsActive(plentyKey) || energyGiven >= maxEnergy {
				return true
			}
			w.char.AddEnergy("beyondthechrysalis-plenty", maxEnergy-energyGiven)
			energyGiven = maxEnergy
			return false
		}, plentyKey)
	}

	w.sequence++
	return false
}
