package main

import (
	"encoding/csv"
	"encoding/json"
	"os"
)

func main() {
	// we want to read a .csv file and get its contents into a struct
	// then we want to write the struct to a json file

	file, err := os.Open("true_abilities.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	abilities := make(map[string]string)

	for i, record := range records {
		if i == 0 {
			continue // skip header row
		}

		if record[1] == "" {
			abilities[record[0]] = "missing_macro"
		} else {
			if record[1] == "no_macro_needed" {
				abilities[record[0]] = "no_macro_needed"
			} else if record[1] == "light_attack" {
				abilities[record[0]] = "&{template:default} {{name=@{name} ❖ Light Attack}} {{Attack=[[1d20+@{focus}]]}} {{Damage=[[2d4+@{power}]]}}"
			} else if record[1] == "heavy_attack" {
				abilities[record[0]] = "&{template:default} {{name=@{name} ❖ Heavy Attack}} {{Attack=[[1d20+@{focus}]]}} {{Damage=[[4d6+2*@{power}]]}}"
			} else if record[1] == "precise_attack" {
				abilities[record[0]] = "&{template:default} {{name=@{name} ❖ Precise Attack}} {{Attack=[[1d20+2*@{focus}]]}} {{Damage=[[2d4+@{power}]]}}"
			} else if record[1] == "light_attack_weapon" {
				abilities[record[0]] = "&{template:default} {{name=@{name} ❖ Light Attack[@{main_weapon_name}]}} {{Attack=[[1d20+@{focus}+@{main_weapon_focus}]]}} {{Damage=[[2d4+@{power}+@{main_weapon_power}]]}}"
			} else if record[1] == "heavy_attack_weapon" {
				abilities[record[0]] = "&{template:default} {{name=@{name} ❖ Heavy Attack[@{main_weapon_name}]}} {{Attack=[[1d20+@{focus}+@{main_weapon_focus}]]}} {{Damage=[[4d6+2*(@{power}+@{main_weapon_power})]]}}"
			} else if record[1] == "precise_attack_weapon" {
				abilities[record[0]] = "&{template:default} {{name=@{name} ❖ Precise Attack[@{main_weapon_name}]}} {{Attack=[[1d20+2*(@{focus}+@{main_weapon_focus})]]}} {{Damage=[[2d4+@{power}+@{main_weapon_power}]]}}"
			} else {
				abilities[record[0]] = record[1]
			}
		}
	}

	bytes, err := json.Marshal(abilities)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("true_abilities_macros.json", bytes, 0644)
	if err != nil {
		panic(err)
	}

}
