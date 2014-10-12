package location

var (
	SESTO UNLocode = "SESTO"
	AUMEL UNLocode = "AUMEL"
	CNHKG UNLocode = "CNHKG"
	USNYC UNLocode = "USNYC"
	USCHI UNLocode = "USCHI"
	JNTKO UNLocode = "JNTKO"
	DEHAM UNLocode = "DEHAM"
)

var (
	Stockholm = Location{SESTO, "Stockholm"}
	Melbourne = Location{AUMEL, "Melbourne"}
	Hongkong  = Location{CNHKG, "Hongkong"}
	NewYork   = Location{USNYC, "New York"}
	Chicago   = Location{USCHI, "Chicago"}
	Tokyo     = Location{JNTKO, "Tokyo"}
	Hamburg   = Location{DEHAM, "Hamburg"}
)
