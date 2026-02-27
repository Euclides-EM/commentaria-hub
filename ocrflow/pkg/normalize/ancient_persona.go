package normalize

import (
	"regexp"
	"strings"

	"github.com/samber/lo"
)

func AncientPersona(name string) []MappedOriginal {
	n := normalizeString(name)
	if n == "" {
		return nil
	}

	contains := func(opts ...string) bool {
		for _, o := range opts {
			if o == "" {
				continue
			}
			if strings.Contains(n, o) {
				return true
			}
		}
		return false
	}

	var out []MappedOriginal

	// helper: add first matching token from a list
	addFirstMatch := func(mapped string, originals ...string) {
		for _, o := range originals {
			if o == "" {
				continue
			}
			if strings.Contains(n, o) {
				out = append(out, MappedOriginal{Original: o, Mapped: mapped})
				return
			}
		}
	}

	// helper: add by regex, capturing matched substring(s) from n
	addByRegex := func(mapped string, re *regexp.Regexp) {
		matches := re.FindAllString(n, -1)
		for _, m := range matches {
			m = strings.TrimSpace(m)
			if m == "" {
				continue
			}
			out = append(out, MappedOriginal{Original: m, Mapped: mapped})
		}
	}

	// Archimedes
	if contains(
		"archimedes",
		"archimede",
		"archimedis",
		"archimede.",
		"darchimedes",
		"archimed",
	) || regexp.MustCompile(`ἀρχιμήδη|αρχιμηδη|archimede`).MatchString(n) {
		addFirstMatch("Archimedes",
			"archimedes",
			"archimede",
			"archimedis",
			"archimede.",
			"darchimedes",
			"archimed",
		)
		addByRegex("Archimedes", regexp.MustCompile(`ἀρχιμήδη|αρχιμηδη|archimede`))
	}

	if contains("avtolyci", "autolyc", "autolycus") {
		addFirstMatch("Autolycus of Pitane", "avtolyci", "autolyc", "autolycus")
	}

	if contains("alexander aphrodiseus", "alexander aphrodis", "aphrodisias") {
		addFirstMatch("Alexander of Aphrodisias", "alexander aphrodiseus", "alexander aphrodis", "aphrodisias")
	}

	if contains("apollonij", "apollonius", "apollonio") {
		addFirstMatch("Apollonius of Perga", "apollonij", "apollonius", "apollonio")
	}

	if contains("aristarchi sami", "aristarchus", "aristarco") {
		addFirstMatch("Aristarchus of Samos", "aristarchi sami", "aristarchus", "aristarco")
	}

	// Aristotle (includes odd false-positive "πλατωνος" from original TS)
	if contains(
		"aristote",
		"aristotele",
		"aristoteleam",
		"aristoteles",
		"aristotelis",
		"daristote",
		"πλατωνος",
	) || regexp.MustCompile(`ἀριστοτε`).MatchString(n) {
		addFirstMatch("Aristotle",
			"aristote",
			"aristotele",
			"aristoteleam",
			"aristoteles",
			"aristotelis",
			"daristote",
			"πλατωνος",
		)
		addByRegex("Aristotle", regexp.MustCompile(`ἀριστοτε`))
	}

	if contains("athenagorae philosophi", "athenagoras") {
		addFirstMatch("Athenagoras of Athens", "athenagorae philosophi", "athenagoras")
	}

	if contains("barlaam") {
		addFirstMatch("Barlaam of Seminara", "barlaam")
	}

	if contains(
		"zamberti",
		"zamberto",
		"bartholomaei zamberti",
		"bartholomæi zamberti",
		"bartholamæi zamberti",
	) {
		addFirstMatch("Bartholomeo Zamberti",
			"bartholomaei zamberti",
			"bartholomæi zamberti",
			"bartholamæi zamberti",
			"zamberti",
			"zamberto",
		)
	}

	if contains("batholomaeo veneto", "batholomæo veneto", "a bartholomæo veneto") {
		addFirstMatch("Bartolomeo Veneto", "a bartholomæo veneto", "batholomaeo veneto", "batholomæo veneto")
	}

	if contains("boetii", "boetij", "boethius", "boetius") {
		addFirstMatch("Boethius", "boethius", "boetius", "boetii", "boetij")
	}

	if contains("boneti latensis", "boni latensis") {
		addFirstMatch("Bonetus Latensis", "boneti latensis", "boni latensis")
	}

	if contains(
		"campane",
		"campani",
		"campani galli transalpini",
		"campani galli",
		"campani ",
		"campano",
		"due tradottioni",
	) {
		addFirstMatch("Campanus of Novara",
			"campani galli transalpini",
			"campani galli",
			"campane",
			"campani",
			"campani ",
			"campano",
			"due tradottioni",
		)
	}

	if contains("candallae", "fr. flussatis candallae", "flussas") {
		addFirstMatch("François de Foix de Candalle", "fr. flussatis candallae", "candallae", "flussas")
	}

	if contains("christophoro clavio", "r.p. christophori clauij", "clavius", "clauij") {
		addFirstMatch("Christopher Clavius", "christophoro clavio", "r.p. christophori clauij", "clavius", "clauij")
	}

	if contains("cleomedes", "cleonidis") {
		addFirstMatch("Cleomedes", "cleomedes", "cleonidis")
	}

	if contains(
		"commandine",
		"federici commandini",
		"fededici commandini",
		"commandini",
	) {
		addFirstMatch("Federico Commandino", "federici commandini", "fededici commandini", "commandine", "commandini")
	}

	if contains("copernican") {
		addFirstMatch("Nicolaus Copernicus", "copernican")
	}

	if contains("galileo", "del galileo", "galilei") {
		addFirstMatch("Galileo Galilei", "del galileo", "galileo", "galilei")
	}

	if contains("torricelli") {
		addFirstMatch("Evangelista Torricelli", "torricelli")
	}

	if contains("eutocij", "eutocius") {
		addFirstMatch("Eutocius of Ascalon", "eutocij", "eutocius")
	}

	if contains("francois viete", "mr. viete", "de lillustre f. viete", "viete") {
		addFirstMatch("François Viète", "de lillustre f. viete", "francois viete", "mr. viete", "viete")
	}

	if contains("fabrice mordente", "mordente") {
		addFirstMatch("Fabrizio Mordente", "fabrice mordente", "mordente")
	}

	if contains("galenus") {
		addFirstMatch("Galen", "galenus")
	}

	if contains("gilberti porretae", "porretae") {
		addFirstMatch("Gilbert de la Porrée", "gilberti porretae", "porretae")
	}

	if contains("henrichvs loritvs glareanvs", "henricvs loritvs glareanvs", "glareanus") {
		addFirstMatch("Henricus Glareanus", "henrichvs loritvs glareanvs", "henricvs loritvs glareanvs", "glareanus")
	}

	// Hero of Alexandria
	if contains("heronis alexandrini") || contains("ηρωνος αλεξανδρεως", "ηρωνος", "αλεξανδρεως") {
		addFirstMatch("Hero of Alexandria", "heronis alexandrini", "ηρωνος αλεξανδρεως", "ηρωνος", "αλεξανδρεως")
	}

	if contains("hypsiclis alexandrini", "hypsiclis", "hypsiclem", "hypsi. alex.") {
		addFirstMatch("Hypsicles of Alexandria", "hypsiclis alexandrini", "hypsiclis", "hypsiclem", "hypsi. alex.")
	}

	if contains("iacobi peletarii cenom.", "peletarii", "peletier") {
		addFirstMatch("Jacques Peletier", "iacobi peletarii cenom.", "peletarii", "peletier")
	}

	if contains("isaaci monachi") {
		addFirstMatch("Isaac Argyros", "isaaci monachi")
	}

	if contains("isidorvm", "isidore") {
		addFirstMatch("Isidore of Seville", "isidorvm", "isidore")
	}

	if contains("ioannis murmelij", "murmelij", "murmelius") {
		addFirstMatch("Johannes Murmellius", "ioannis murmelij", "murmelij", "murmelius")
	}

	if contains("john dee", "m. i. dee", "i. dee", "dee of london") {
		addFirstMatch("John Dee", "dee of london", "john dee", "m. i. dee", "i. dee")
	}

	if contains("marinus", "marini dialectici") {
		addFirstMatch("Marinus of Neapolis", "marini dialectici", "marinus")
	}

	if contains("martianvs rota") {
		addFirstMatch("Martianus Rota", "martianvs rota")
	}

	if contains("maurolyci", "mavrolyci", "maurolico") {
		addFirstMatch("Francesco Maurolico", "maurolyci", "mavrolyci", "maurolico")
	}

	if contains("menelai", "menelaus") {
		addFirstMatch("Menelaus of Alexandria", "menelai", "menelaus")
	}

	if contains("nicephori", "nicephorus") {
		addFirstMatch("Nicephorus", "nicephori", "nicephorus")
	}

	if contains("procli", "proclus", "πρόκλου") {
		addFirstMatch("Proclus", "πρόκλου", "procli", "proclus")
	}

	if contains("pappi mechanici", "pappi", "pappus") {
		addFirstMatch("Pappus of Alexandria", "pappi mechanici", "pappi", "pappus")
	}

	if contains("platone", "platus", "πλάτων", "πλατων", "πλάτωνος", "plato", "γλάπτων") {
		addFirstMatch("Plato", "πλάτωνος", "πλάτων", "πλατων", "platone", "platus", "plato", "γλάπτων")
	}

	if contains("pythagorean", "pytagorean", "πυθαγόρας", "γυπαγόρας") {
		addFirstMatch("Pythagorean", "pythagorean", "pytagorean", "πυθαγόρας", "γυπαγόρας")
	}

	if contains("robert hves", "robert hues") {
		addFirstMatch("Robert Hues", "robert hves", "robert hues")
	}

	if contains("rhazes") {
		addFirstMatch("Abu Bakr al-Razi", "rhazes")
	}

	if contains("rodolphi agricolae") {
		addFirstMatch("Rodolphus Agricola", "rodolphi agricolae")
	}

	if contains("stevin") {
		addFirstMatch("Simon Stevin", "stevin")
	}

	if contains("sacrobosco") {
		addFirstMatch("Johannes de Sacrobosco", "sacrobosco")
	}

	if contains("scipio vegius") {
		addFirstMatch("Scipione Vizzani", "scipio vegius")
	}

	if contains("theodosii", "theodosij") {
		addFirstMatch("Theodosius of Bithynia", "theodosii", "theodosij")
	}

	if contains("theonis alexandrini", "theonis", "theon", "θεωνος", "θεῶνος", "θέωνος") {
		addFirstMatch("Theon of Alexandria", "theonis alexandrini", "theonis", "theon", "θεωνος", "θεῶνος", "θέωνος")
	}

	if contains("timeus", "timaeus") {
		addFirstMatch("Timaeus of Locri", "timeus", "timaeus")
	}

	if contains("zamber", "due tradottioni") {
		addFirstMatch("Bartholomeo Zamberti", "zamber", "due tradottioni")
	}

	// final dedupe: (Original, Mapped)
	out = lo.UniqBy(out, func(x MappedOriginal) string {
		return x.Original + "\x00" + x.Mapped
	})

	return out
}
