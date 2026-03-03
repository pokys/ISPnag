package mood

import (
	"math/rand"
	"time"
)

var personalityByMood = map[string][]string{
	"🙂 klidná": {
		"Síť dnes šlape jak hodinky.",
		"Klid na všech frekvencích.",
		"Dneska je síť až podezřele hodná.",
		"Bez bouřek, bez dramat.",
		"Dnešní provoz je nuda v tom nejlepším slova smyslu.",
		"Síť je dnes v režimu pohoda.",
		"Monitoring by se dnes skoro nudil.",
		"Vypadá to na den bez překvapení.",
	},
	"😐 lehce nervózní": {
		"Síť si dnes trochu postěžovala.",
		"Nic nehoří, ale pár kabelů se mračí.",
		"Lehké poryvy nervozity, zatím bez paniky.",
		"Lehká oblačnost nad několika uzly.",
		"Síť má drobné nálady, nic zásadního.",
	},
	"😬 mrzutá": {
		"Dnes to není zen, spíš lehký chaos.",
		"Oblačno s občasným zaburácením.",
		"Nad sítí je zataženo, ale bez lijáku.",
	},
	"⚠️ nestabilní": {
		"Deštník a kafe doporučeno.",
		"Síť dneska promoká po částech.",
	},
	"🔥 naštvaná": {
		"Dneska to voní incidentem.",
		"Tohle se samo nespraví, bohužel.",
		"Hrom už je slyšet, teď jen kdy uhodí.",
		"Výstraha platí: dnes to bude chtít rychlou reakci.",
		"Bude to rachot, držte se.",
	},
}

func FromAnnoyance(ann int) string {
	switch {
	case ann <= 20:
		return "🙂 klidná"
	case ann <= 50:
		return "😐 lehce nervózní"
	case ann <= 80:
		return "😬 mrzutá"
	case ann <= 120:
		return "⚠️ nestabilní"
	default:
		return "🔥 naštvaná"
	}
}

func RandomPersonalityLine(mood string, seed int64) string {
	choices := personalityByMood[mood]
	if len(choices) == 0 {
		return ""
	}
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	r := rand.New(rand.NewSource(seed))
	return choices[r.Intn(len(choices))]
}
