package normalize

import (
	"regexp"
	"strings"
)

func AncientPersona(name string) string {
	n := String(name)
	if n == "" {
		return ""
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

	// Archimedes
	if contains(
		"archimedes",
		"archimede",
		"archimedis",
		"archimede.",
		"darchimedes",
		"archimed",
	) || regexp.MustCompile(`ἀρχιμήδη|αρχιμηδη|archimede`).MatchString(n) {
		return "Archimedes"
	}

	if contains("avtolyci", "autolyc", "autolycus") {
		return "Autolycus of Pitane"
	}

	if contains("alexander aphrodiseus", "alexander aphrodis", "aphrodisias") {
		return "Alexander of Aphrodisias"
	}

	if contains("apollonij", "apollonius", "apollonio") {
		return "Apollonius of Perga"
	}

	if contains("aristarchi sami", "aristarchus", "aristarco") {
		return "Aristarchus of Samos"
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
		return "Aristotle"
	}

	if contains("athenagorae philosophi", "athenagoras") {
		return "Athenagoras of Athens"
	}

	if contains("barlaam") {
		return "Barlaam of Seminara"
	}

	if contains(
		"zamberti",
		"zamberto",
		"bartholomaei zamberti",
		"bartholomæi zamberti",
		"bartholamæi zamberti",
	) {
		return "Bartholomeo Zamberti"
	}

	if contains("batholomaeo veneto", "batholomæo veneto", "a bartholomæo veneto") {
		return "Bartolomeo Veneto"
	}

	if contains("boetii", "boetij", "boethius", "boetius") {
		return "Boethius"
	}

	if contains("boneti latensis", "boni latensis") {
		return "Bonetus Latensis"
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
		return "Campanus of Novara"
	}

	if contains("candallae", "fr. flussatis candallae", "flussas") {
		return "François de Foix de Candalle"
	}

	if contains("christophoro clavio", "r.p. christophori clauij", "clavius", "clauij") {
		return "Christopher Clavius"
	}

	if contains("cleomedes", "cleonidis") {
		return "Cleomedes"
	}

	if contains(
		"commandine",
		"federici commandini",
		"fededici commandini",
		"commandini",
	) {
		return "Federico Commandino"
	}

	if contains("copernican") {
		return "Nicolaus Copernicus"
	}

	if contains("galileo", "del galileo", "galilei") {
		return "Galileo Galilei"
	}

	if contains("torricelli") {
		return "Evangelista Torricelli"
	}

	if contains("eutocij", "eutocius") {
		return "Eutocius of Ascalon"
	}

	if contains("francois viete", "mr. viete", "de lillustre f. viete", "viete", "viete") {
		return "François Viète"
	}

	if contains("fabrice mordente", "mordente") {
		return "Fabrizio Mordente"
	}

	if contains("galenus") {
		return "Galen"
	}

	if contains("gilberti porretae", "porretae") {
		return "Gilbert de la Porrée"
	}

	if contains("henrichvs loritvs glareanvs", "henricvs loritvs glareanvs", "glareanus") {
		return "Henricus Glareanus"
	}

	// Hero of Alexandria
	if contains("heronis alexandrini") || contains("ηρωνος αλεξανδρεως", "ηρωνος", "αλεξανδρεως") {
		return "Hero of Alexandria"
	}

	if contains("hypsiclis alexandrini", "hypsiclis", "hypsiclem", "hypsi. alex.") {
		return "Hypsicles of Alexandria"
	}

	if contains("iacobi peletarii cenom.", "peletarii", "peletier") {
		return "Jacques Peletier"
	}

	if contains("isaaci monachi") {
		return "Isaac Argyros"
	}

	if contains("isidorvm", "isidore") {
		return "Isidore of Seville"
	}

	if contains("ioannis murmelij", "murmelij", "murmelius") {
		return "Johannes Murmellius"
	}

	if contains("john dee", "m. i. dee", "i. dee", "dee of london") {
		return "John Dee"
	}

	if contains("marinus", "marini dialectici") {
		return "Marinus of Neapolis"
	}

	if contains("martianvs rota") {
		return "Martianus Rota"
	}

	if contains("maurolyci", "mavrolyci", "maurolico") {
		return "Francesco Maurolico"
	}

	if contains("menelai", "menelaus") {
		return "Menelaus of Alexandria"
	}

	if contains("nicephori", "nicephorus") {
		return "Nicephorus"
	}

	if contains("procli", "proclus", "πρόκλου") {
		return "Proclus"
	}

	if contains("pappi mechanici", "pappi", "pappus") {
		return "Pappus of Alexandria"
	}

	if contains("platone", "platus", "πλάτων", "πλατων", "πλάτωνος", "plato", "γλάπτων") {
		return "Plato"
	}

	if contains("pythagorean", "pytagorean", "πυθαγόρας", "γυπαγόρας") {
		return "Pythagoras"
	}

	if contains("robert hves", "robert hues") {
		return "Robert Hues"
	}

	if contains("rhazes") {
		return "Abu Bakr al-Razi"
	}

	if contains("rodolphi agricolae") {
		return "Rodolphus Agricola"
	}

	if contains("stevin") {
		return "Simon Stevin"
	}

	if contains("sacrobosco") {
		return "Johannes de Sacrobosco"
	}

	if contains("scipio vegius") {
		return "Scipione Vizzani"
	}

	if contains("theodosii", "theodosij") {
		return "Theodosius of Bithynia"
	}

	if contains("theonis alexandrini", "theonis", "theon", "θεωνος", "θεῶνος", "θέωνος") {
		return "Theon of Alexandria"
	}

	if contains("timeus", "timaeus") {
		return "Timaeus of Locri"
	}

	if contains("zamber", "due tradottioni") {
		return "Bartholomeo Zamberti"
	}

	return ""
}
