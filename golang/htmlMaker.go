package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"sort"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

var otherCraftingAbilities []string = []string{
	"detailed_finish",
	"light_work",
	"siphon_material",
	"repair_armor",
	"thermal_quench",
	"repair_artifice",
	"overclock",
	"repair_weapon",
	"steady_craft_ii",
	"memory_of_the_grove",
	"stalks_of_comparison",
	"repair_wood_weapon",
	"in_the_zone",
	"perfect_seasoning",
	"craft_pemmican",
	"four_field_system",
	"focused_detonation",
	"take_it_easy",
	"pull_the_weeds",
	"verdant_instinct",
	"divining_petalfall",
	"inspired_finish",
	"expertise_study_(armorsmithing)",
	"expertise_study_(blacksmithing)",
	"expertise_study_(artificing)",
	"expertise_study_(carpentry)",
	"craft_artificers_glove",
	"reverse_polarity",
	"grindstone_echo",
	"power_rock_strike",
	"efficient_strike",
	"stable_foundation",
	"deconstruction",
}

type ClassInfo struct {
	Name string
	ID   string
	Tier int
}

type KeyAbility struct {
	ID       string
	ClassID  []string
	Name     string
	Benefits []string
}

type TrueAbility struct {
	ID            string
	ClassID       []string
	LevelRequired int
	Name          string
	Keywords      string
	Range         string
	Requirement   string
	Description   string
	RPcost        string
	APcost        string
	MPcost        string
	Othercost     string
}

type CraftingAbility struct {
	ID          string
	Name        string
	Keywords    string
	Cost        string
	Description string
}

type Breakthrough struct {
	ID           string
	Name         string
	Cost         string
	Requirement  string
	Requirements string
	Description  string
}

type Race struct {
	ID   string
	Name string
}

type Subrace struct {
	ID   string
	Name string
}

func main() {
	breakthroughsMap := make(map[string]Breakthrough)
	trueAbilities := make(map[string]TrueAbility)
	craftingAbilities := make(map[string]CraftingAbility)
	racesAbilities := make(map[string]TrueAbility)
	keyAbilities := make(map[string]KeyAbility)
	classesMap := make(map[string]ClassInfo)
	racesMap := make(map[string]string)
	subracesMap := make(map[string]string)

	// Loading all JSONS

	bytes, err := os.ReadFile("./webscrapper/breakthroughs.json")
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(bytes, &breakthroughsMap); err != nil {
		panic(err)
	}

	bytes, err = os.ReadFile("./webscrapper/true_abilities.json")
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(bytes, &trueAbilities); err != nil {
		panic(err)
	}

	bytes, err = os.ReadFile("./webscrapper/crafting_abilities.json")
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(bytes, &craftingAbilities); err != nil {
		panic(err)
	}

	bytes, err = os.ReadFile("./webscrapper/key_abilities.json")
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(bytes, &keyAbilities); err != nil {
		panic(err)
	}

	bytes, err = os.ReadFile("./webscrapper/classes.json")
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(bytes, &classesMap); err != nil {
		panic(err)
	}

	bytes, err = os.ReadFile("./webscrapper/subraces.json")
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(bytes, &subracesMap); err != nil {
		panic(err)
	}

	bytes, err = os.ReadFile("./webscrapper/races.json")
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(bytes, &racesMap); err != nil {
		panic(err)
	}

	bytes, err = os.ReadFile("./webscrapper/races_true_abilities.json")
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(bytes, &racesAbilities); err != nil {
		panic(err)
	}

	newTrueAbilities := make(map[string]TrueAbility)
	maps.Copy(newTrueAbilities, trueAbilities)

	// now we must "FIX" the crafting abilities that are missing from the webscrapper
	for _, id := range otherCraftingAbilities {
		if _, ok := trueAbilities[id]; ok {
			delete(newTrueAbilities, id)
			var cost string = ""
			if trueAbilities[id].APcost != "" {
				cost = trueAbilities[id].APcost
			}
			if trueAbilities[id].MPcost != "" {
				if cost != "" {
					cost += ", "
				}
				cost += trueAbilities[id].MPcost
			}
			if trueAbilities[id].RPcost != "" {
				if cost != "" {
					cost += ", "
				}
				cost += trueAbilities[id].RPcost
			}
			if trueAbilities[id].Othercost != "" {
				if cost != "" {
					cost += ", "
				}
				cost += trueAbilities[id].Othercost
			}
			if cost == "" {
				cost = "--"
			}

			craftingAbilities[id] = CraftingAbility{
				ID:          trueAbilities[id].ID,
				Name:        trueAbilities[id].Name,
				Keywords:    trueAbilities[id].Keywords,
				Description: trueAbilities[id].Description,
				Cost:        cost,
			}
		} else {
			fmt.Println("Crafting ability not found as true ability:", id)
		}
	}

	trueAbilities = newTrueAbilities

	// Create the HTML file
	htmlFile, err := os.Create("update-data.html")
	if err != nil {
		panic(err)
	}
	defer htmlFile.Close()

	// ============================= RACES START =============================

	var races []Race
	for raceID, race := range racesMap {
		r := Race{
			ID:   raceID,
			Name: race,
		}
		races = append(races, r)
		fmt.Printf("id: %s, name: %s\n", r.ID, r.Name)
	}

	sort.Slice(races, func(i, j int) bool {
		return races[i].ID < races[j].ID
	})

	fmt.Fprintln(htmlFile, "<!-- All RACES Picker List START -->")

	fmt.Fprintln(htmlFile, `<option value="">--Choose--</option>`)
	for _, race := range races {
		fmt.Fprintf(htmlFile, `<option value="%s">%s</option>`, race.ID, race.Name)
	}
	fmt.Fprintln(htmlFile, "<!-- All RACES Picker List END -->")

	// ============================= RACES END =============================

	// ============================= SUBRACES START =============================

	var subraces []Subrace
	for subraceID, subrace := range subracesMap {
		sr := Subrace{
			ID:   subraceID,
			Name: subrace,
		}
		subraces = append(subraces, sr)
		fmt.Printf("id: %s, name: %s\n", sr.ID, sr.Name)
	}

	sort.Slice(subraces, func(i, j int) bool {
		return subraces[i].ID < subraces[j].ID
	})

	fmt.Fprintln(htmlFile, "<!-- All SUBRACES Picker List START -->")

	fmt.Fprintln(htmlFile, `<option value="">--Choose--</option>`)
	for _, subrace := range subraces {
		fmt.Fprintf(htmlFile, `<option value="%s">%s</option>`, subrace.ID, subrace.Name)
	}
	fmt.Fprintln(htmlFile, "<!-- All SUBRACES Picker List END -->")

	// ============================= SUBRACES END =============================

	// ============================= CLASSES START =============================

	var classes []ClassInfo

	for _, class := range classesMap {

		c := ClassInfo{
			ID:   class.ID,
			Name: class.Name,
			Tier: class.Tier,
		}
		classes = append(classes, c)
		fmt.Printf("id: %s, name: %s, tier: %d\n", c.ID, c.Name, c.Tier)
	}

	sort.Slice(classes, func(i, j int) bool {
		return classes[i].ID < classes[j].ID
	})

	fmt.Fprintln(htmlFile, "<!-- All Clases Picker List START -->")

	fmt.Fprintln(htmlFile, `<option value="">--Select--</option>`)
	fmt.Fprintln(htmlFile, `<option value="custom">Custom</option>`)
	for _, class := range classes {
		fmt.Fprintf(htmlFile, `<option value="%s">%s</option>`, class.ID, class.Name)
	}
	fmt.Fprintln(htmlFile, "<!-- All Clases Picker List END -->")

	// ============================= CLASSES END =============================
	// ============================= BREAKTHROUGHS =============================

	var breakthroughs []Breakthrough

	for _, bt := range breakthroughsMap {

		if bt.ID[0] >= '0' && bt.ID[0] <= '9' {
			bt.ID = "x" + bt.ID
		}

		b := Breakthrough{
			ID:           bt.ID,
			Name:         bt.Name,
			Cost:         bt.Cost,
			Requirements: strings.ReplaceAll(bt.Requirement, `"`, "'"),
			Description:  strings.ReplaceAll(bt.Description, `"`, "'"),
		}
		breakthroughs = append(breakthroughs, b)
		fmt.Printf("id: %s, name: %s, cost: %s\n", b.ID, b.Name, b.Cost)
	}
	sort.Slice(breakthroughs, func(i, j int) bool {
		return breakthroughs[i].Name < breakthroughs[j].Name
	})

	fmt.Fprintln(htmlFile, "<!-- All Breakthroughs Picker List START -->")

	fmt.Fprintln(htmlFile, `<option value="">--Select--</option>`)
	fmt.Fprintln(htmlFile, `<option value="custom">Custom</option>`)

	for _, breakthrough := range breakthroughs {
		fmt.Fprintf(htmlFile, `<option value="%s">%s</option>`, breakthrough.ID, breakthrough.Name)
	}

	fmt.Fprintln(htmlFile, "<!-- All Breakthroughs Picker List END -->")

	// ============================= BREAKTHROUGHS END =============================

	// ============================= ABILITIES START =============================

	type Ability struct {
		ID            string
		ClassID       []string
		LevelRequired int
		Name          string
		Type          string
		Keywords      string
		Range         string
		Description   string
		Requirements  string
		Costs         string
		OtherCosts    string
		Benefits      string
		AtkTypes      string
	}

	var keyAbilitiesArray []Ability
	var regularAbilitiesArray []Ability
	var craftingAbilitiesArray []Ability

	for _, ab := range craftingAbilities {
		ability := Ability{
			ID:          ab.ID,
			Name:        ab.Name,
			Type:        "crafting_ability",
			Description: strings.ReplaceAll(ab.Description, `"`, "'"),
			Costs:       strings.ReplaceAll(ab.Cost, `"`, "'"),
		}

		craftingAbilitiesArray = append(craftingAbilitiesArray, ability)
	}
	sort.Slice(craftingAbilitiesArray, func(i, j int) bool {
		return craftingAbilitiesArray[i].ID < craftingAbilitiesArray[j].ID
	})

	for _, ab := range trueAbilities {
		var costs string = ""
		if ab.APcost != "" {
			costs += ab.APcost
		}
		if ab.MPcost != "" {
			if costs != "" {
				costs += ", "
			}
			costs += ab.MPcost
		}
		if ab.RPcost != "" {
			if costs != "" {
				costs += ", "
			}
			costs += ab.RPcost
		}

		ability := Ability{
			ID:            ab.ID,
			ClassID:       ab.ClassID,
			LevelRequired: ab.LevelRequired,
			Name:          ab.Name,
			Type:          "true_ability",
			Keywords:      ab.Keywords,
			Range:         ab.Range,
			Description:   strings.TrimSpace(strings.ReplaceAll(ab.Description, `"`, "'")),
			Requirements:  strings.ReplaceAll(ab.Requirement, `"`, "'"),
			Costs:         strings.ReplaceAll(costs, `"`, "'"),
		}
		regularAbilitiesArray = append(regularAbilitiesArray, ability)
	}
	for _, ab := range racesAbilities {
		var costs string = ""
		if ab.APcost != "" {
			costs += ab.APcost
		}
		if ab.MPcost != "" {
			if costs != "" {
				costs += ", "
			}
			costs += ab.MPcost
		}
		if ab.RPcost != "" {
			if costs != "" {
				costs += ", "
			}
			costs += ab.RPcost
		}

		ability := Ability{
			ID:           ab.ID,
			ClassID:      []string{},
			Name:         ab.Name,
			Type:         "true_ability",
			Keywords:     ab.Keywords,
			Range:        ab.Range,
			Description:  strings.TrimSpace(strings.ReplaceAll(ab.Description, `"`, "'")),
			Requirements: strings.ReplaceAll(ab.Requirement, `"`, "'"),
			Costs:        strings.ReplaceAll(costs, `"`, "'"),
		}

		found := false
		for _, a := range regularAbilitiesArray {
			if a.ID == ability.ID {
				found = true
				break
			}
		}
		if found {
			continue
		}

		regularAbilitiesArray = append(regularAbilitiesArray, ability)
	}

	for _, ab := range keyAbilities {
		benefits := strings.Join(ab.Benefits, " ")
		benefits2 := fmt.Sprintf(`%s`, benefits)

		ability := Ability{
			ID:       ab.ID,
			ClassID:  ab.ClassID,
			Name:     ab.Name,
			Type:     "key_ability",
			Benefits: strings.TrimSpace(strings.ReplaceAll(benefits2, `"`, "'")),
		}
		keyAbilitiesArray = append(keyAbilitiesArray, ability)
	}

	sort.Slice(regularAbilitiesArray, func(i, j int) bool {
		return regularAbilitiesArray[i].ID < regularAbilitiesArray[j].ID
	})
	sort.Slice(keyAbilitiesArray, func(i, j int) bool {
		return keyAbilitiesArray[i].ID < keyAbilitiesArray[j].ID
	})

	fmt.Fprintln(htmlFile, "<!-- All Key Abilities Picker List START -->")

	fmt.Fprintln(htmlFile, "<!-- All Key abilities Picker List END -->")

	for _, ability := range keyAbilitiesArray {
		fmt.Fprintf(htmlFile, `<option value="%s">%s</option>`, ability.ID, ability.Name)
	}

	fmt.Fprintln(htmlFile, "<!-- All Abilities Picker List START -->")

	fmt.Fprintln(htmlFile, `<option value="">--Select--</option>`)
	fmt.Fprintln(htmlFile, `<option value="custom">Custom</option>`)

	for _, ability := range regularAbilitiesArray {
		fmt.Fprintf(htmlFile, `<option value="%s">%s</option>`, ability.ID, ability.Name)
	}

	fmt.Fprintln(htmlFile, "<!-- All abilities Picker List END -->")

	// Crafting abilities
	fmt.Fprintln(htmlFile, "<!-- All Crafting Abilities Picker List START -->")

	fmt.Fprintln(htmlFile, `<option value="">--Select--</option>`)
	fmt.Fprintln(htmlFile, `<option value="custom">Custom</option>`)

	for _, ability := range craftingAbilitiesArray {
		fmt.Fprintf(htmlFile, `<option value="%s">%s</option>`, ability.ID, ability.Name)
	}

	fmt.Fprintln(htmlFile, "<!-- All Crafting abilities Picker List END -->")

	// ============================= ABILITIES END =============================

	fmt.Fprintln(htmlFile, "<!-- SCRIPTS TO REPLACE START -->")

	fmt.Fprintln(htmlFile, "<script>\n")

	// Classes array
	var arraystring string
	arraystring = "{"
	for i, class := range classes {
		if i > 0 {
			arraystring += ",\n"
		}
		arraystring += fmt.Sprintf(`"%s": {name: "%s",tier:%d}`, class.ID, class.Name, class.Tier)
	}
	arraystring += "}"

	fmt.Fprintf(htmlFile, `const classList =%s;`, arraystring)
	fmt.Fprintln(htmlFile, "\n")

	// Breakthroughs array
	var arraystring2 string
	arraystring2 = "{"
	for i, breakthrough := range breakthroughs {
		if i > 0 {
			arraystring2 += ",\n"
		}
		description := strings.Replace(breakthrough.Description, "\n", " ", -1)
		arraystring2 += fmt.Sprintf(`"%s": {name: "%s",cost:%s,requirements:"%s",description:"%s"}`, breakthrough.ID, breakthrough.Name, breakthrough.Cost, breakthrough.Requirements, description)
		// if i > 0 {
		// 	arraystring2 += "\n"
		// }
	}
	arraystring2 += "}"

	fmt.Fprintf(htmlFile, `const breakthroughList =%s;`, arraystring2)
	fmt.Fprintln(htmlFile, "\n")

	// Key Abilities array
	var arraystring6 string
	arraystring6 = "{"
	for i, ability := range keyAbilitiesArray {
		if i > 0 {
			arraystring6 += ",\n"
		}
		description := strings.Replace(ability.Description, "\n", " ", -1)
		classesArray := "["
		for j, classID := range ability.ClassID {
			if j > 0 {
				classesArray += ","
			}
			classesArray += `"` + classID + `"`
		}
		classesArray += "]"
		arraystring6 += fmt.Sprintf(`"%s": {name: "%s", class:%v,description:"%s",benefits:"%s"}`, ability.ID, ability.Name, classesArray, description, ability.Benefits)
	}
	arraystring6 += "}"

	fmt.Fprintf(htmlFile, `const keyAbilityList =%s;`, arraystring6)
	fmt.Fprintln(htmlFile, "\n")

	// Abilities array
	var arraystring3 string
	arraystring3 = "{"
	for i, ability := range regularAbilitiesArray {
		if i > 0 {
			arraystring3 += ",\n"
		}
		description := strings.Replace(ability.Description, "\n", " ", -1)
		classesArray := "["
		for j, classID := range ability.ClassID {
			if j > 0 {
				classesArray += ","
			}
			classesArray += `"` + classID + `"`
		}
		classesArray += "]"
		arraystring3 += fmt.Sprintf(`"%s": {name: "%s", class:%v, level:"%d",type:"%s",keywords:"%s",range:"%s",description:"%s",requirements:"%s",costs:"%s",benefits:"%s"}`, ability.ID, ability.Name, classesArray, ability.LevelRequired, ability.Type, ability.Keywords, ability.Range, description, ability.Requirements, ability.Costs, ability.Benefits)
	}
	arraystring3 += "}"

	fmt.Fprintf(htmlFile, `const abilityList =%s;`, arraystring3)
	fmt.Fprintln(htmlFile, "\n")

	// Abilities macros map
	var arraystring5 string
	bytes, err = os.ReadFile("csvfetcher/true_abilities_macros.json")
	if err != nil {
		panic(err)
	}

	abilityMMap := make(map[string]string)
	if err = json.Unmarshal(bytes, &abilityMMap); err != nil {
		panic(err)
	}

	spew.Dump(abilityMMap)
	arraystring5 = "{"
	i := 0
	for id, macro := range abilityMMap {
		if i > 0 {
			arraystring5 += ",\n"
		}
		arraystring5 += fmt.Sprintf(`"%s": "%s"`, id, strings.ReplaceAll(macro, `"`, `\"`))
		i++
	}
	arraystring5 += "}"

	fmt.Fprintf(htmlFile, `const abilityMacroMap =%s;`, arraystring5)
	fmt.Fprintln(htmlFile, "\n")

	// Crafting Abilities array
	var arraystring4 string
	arraystring4 = "{"
	for i, ability := range craftingAbilitiesArray {
		if i > 0 {
			arraystring4 += ",\n"
		}
		description := strings.Replace(ability.Description, "\n", " ", -1)
		arraystring4 += fmt.Sprintf(`"%s": {name: "%s",type:"%s",description:"%s",requirements:"%s",costs:"%s",othercosts:"%s"}`, ability.ID, ability.Name, ability.Type, description, ability.Requirements, ability.Costs, ability.OtherCosts)
	}
	arraystring4 += "}"

	fmt.Fprintf(htmlFile, `const craftingAbilityList =%s;`, arraystring4)
	fmt.Fprintln(htmlFile, "\n")

	// End of scripts to replace

	fmt.Fprintln(htmlFile, "</script>\n")
	fmt.Fprintln(htmlFile, `<!-- SCRIPTS TO REPLACE END -->`)
}
