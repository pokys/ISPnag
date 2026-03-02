package mood

import (
	"math/rand"
	"time"
)

var personalityByMood = map[string][]string{
	"🙂 klidná": {
		"Všude klid. Podezřele velký klid.",
		"Síť se dnes chová ukázkově.",
		"Síť je dnes v neobvyklém klidu.",
		"Dnes to běží bez dramat.",
		"Všechno drží pohromadě, jak má.",
		"Dnes síť vypadá civilizovaně.",
		"Ráno bez překvapení. Zatím.",
		"Dnešek začíná nečekaně klidně.",
	},
	"😐 lehce nervózní": {
		"Zachycené drobné potíže.",
		"Nic urgentního, jen lehké brblání.",
		"Síť si dnes trochu postěžovala.",
		"Menší výkyvy, ale nic zásadního.",
		"Pár míst je dnes citlivějších.",
		"Zatím jen drobné šumy.",
		"Něco se děje, ale bez paniky.",
		"Lehké turbulence v provozu.",
	},
	"😬 mrzutá": {
		"Některé části sítě vstaly levou nohou.",
		"Situace začíná být zajímavá.",
		"Několik zařízení si říká o kontrolu.",
		"Dnes to není úplně bez práce.",
		"Síť je místy nevrlá.",
		"Klid to není, ale dá se to chytit včas.",
		"Některé linky začínají protestovat.",
		"Tady už je dobré zbystřit.",
	},
	"⚠️ nestabilní": {
		"Je cítit předincidentní energie.",
		"Než se do toho pustíte, dejte si kávu.",
		"Tohle už chce aktivní dohled.",
		"Síť dává jasné varovné signály.",
		"Dneska to chce rychlou kontrolu kritických bodů.",
		"Riziko eskalace je dnes vyšší.",
		"Něco se láme, je čas to podchytit.",
		"Bude lepší jednat dřív než později.",
	},
	"🔥 naštvaná": {
		"Uživatelé si toho brzy všimnou.",
		"Tohle skončí poradou.",
		"Tady už hrozí, že to přeroste v incident.",
		"Dnes to bude chtít rychlý zásah.",
		"Síť je dnes opravdu ve špatné náladě.",
		"Tohle už je na prioritu číslo jedna.",
		"Bez akce se to samo nesrovná.",
		"Dneska to bude bolet, pokud to necháme být.",
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
