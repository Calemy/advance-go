package main

type Mod struct { //We do not support modifiers. I don't care. - v3 (Nanoo)
	Acronym string `json:"acronym"`
}

func convertMods(mods []Mod) string {
	if len(mods) == 0 {
		return "{}"
	}

	out := "{"
	for i, m := range mods {
		out += m.Acronym
		if i < len(mods)-1 {
			out += ","
		}
	}
	out += "}"
	return out
}
