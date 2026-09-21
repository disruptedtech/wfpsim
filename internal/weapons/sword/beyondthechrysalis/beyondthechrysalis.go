func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
    w := &Weapon{
        char:   char,
        core:   c,
        refine: p.Refine,
    }

    // Fixed Refinement Math: (w.refine - 1) ensures R1 yields exactly the base values
    cdBuff := 0.40 + float64(w.refine-1)*0.16
    swirlBuff := 0.27 + float64(w.refine-1)*0.09
    maxEnergy := 4.5 + float64(w.refine-1)*0.5

    m := make([]float64, attributes.EndStatType)

    char.AddStatMod(character.StatMod{
        Base:         modifier.NewBase("beyondthechrysalis", -1),
        AffectedStat: attributes.NoStat,
        Amount: func() []float64 {
            m[attributes.CD] = 0
            // Fixed Attribute: Replaced AnemoP with the WFPSIM specific reaction attribute
            m[attributes.StellarSwirlP] = 0 
            
            if char.StatusIsActive(devotionKey) {
                m[attributes.CD] = cdBuff
            }
            if char.StatusIsActive(defianceKey) {
                m[attributes.StellarSwirlP] = swirlBuff
            }
            return m
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
            // Fixed Trigger: Added the missing status application for the reaction buff
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
