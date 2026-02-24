package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/davecgh/go-spew/spew"
)

const baseURL = "https://rpg.angelssword.com/game/latest"

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
	RaceID        []string
	SubraceID     []string
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
type CraftingAbilityRaw struct {
	ID   string
	Name string
	Text string
}

type Breakthrough struct {
	ID          string
	Name        string
	Cost        string
	Requirement string
	Description string
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
	getBreakthroughs()
	getClassesAndAbilities()
	getRacesAndAbilities()
}

func stringBeforeAfter(s string, sep string) (string, string, bool) {
	i := strings.Index(s, sep)
	if i != -1 {
		before := s[:i]
		after := s[i+len(sep):]
		// fmt.Println("Before:", before)
		// fmt.Println("After:", after)
		return before, after, true
	}
	return s, s, false
}

// func printNCharacters(s string, n int) string {
// 	if len(s) <= n {
// 		return fmt.Sprintln(s)
// 	} else {
// 		return fmt.Sprintln(s[:n] + "...")
// 	}
// }

func getBreakthroughs() {
	fmt.Println("Getting Breakthroughs...")
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var abilitySegments []string
	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL+"/breakthroughs"),
		chromedp.WaitVisible(`mat-expansion-panel`, chromedp.ByQuery),
		chromedp.Sleep(10*time.Second),
		chromedp.Evaluate(` Array.from(document.querySelectorAll('mat-expansion-panel')).map(panel => panel.outerHTML) `, &abilitySegments),
	)
	if err != nil {
		fmt.Println("Error getting hrefs:", err)
		return
	}
	breakthroughs := make(map[string]Breakthrough, 0)

	for _, html := range abilitySegments {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			fmt.Println("Error parsing segment:", err)
			continue
		}

		var breakthrough Breakthrough

		rawName := doc.Find("mat-panel-title span").Text()

		abilityName := strings.Split(rawName, "○ ")[1]
		abilityID := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(abilityName, " ", "_"), "'", ""))
		breakthrough.ID = abilityID
		breakthrough.Name = abilityName

		// fmt.Printf("Breakthrough %d: %s [%s]\n", i+1, abilityName, abilityID)

		doc.Find("app-breakthrough mat-card-content li").Each(func(j int, s *goquery.Selection) {
			text := s.Text()
			if j == 0 {
				arr := strings.Split(text, "Cost")
				breakthrough.Cost = strings.TrimSpace(arr[1])
			} else if j == 1 {
				arr := strings.Split(text, "Requirements")
				breakthrough.Requirement = strings.TrimSpace(arr[1])
			}
		})
		doc.Find("app-breakthrough mat-card-content .description").Each(func(j int, s *goquery.Selection) {
			text := s.Text()
			breakthrough.Description = strings.TrimSpace(text)
		})

		breakthroughs[abilityID] = breakthrough
	}

	breakthroughsJSON, err := json.MarshalIndent(breakthroughs, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling breakthroughs to JSON:", err)
		return
	}

	err = os.WriteFile("breakthroughs.json", breakthroughsJSON, 0644)
	if err != nil {
		fmt.Println("Error writing breakthroughs.json file:", err)
		return
	}
	fmt.Printf("Wrote %d breakthroughs to breakthroughs.json\n", len(breakthroughs))
}

func getClassesAndAbilities() {
	levelAbilities := []int{1, 2, 4, 6, 8}
	fmt.Println("Getting Classes and Abilities...")
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var hrefs []string
	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL+"/classes"),
		chromedp.WaitVisible(`app-class-card`, chromedp.ByQuery),
		chromedp.Sleep(10*time.Second),
		chromedp.Evaluate(`
            Array.from(document.querySelectorAll('app-class-card a')).map(a => a.href)
        `, &hrefs),
	)
	if err != nil {
		fmt.Println("Error getting hrefs:", err)
		return
	}

	classes := make(map[string]ClassInfo, 0)
	keyAbilities := make(map[string]KeyAbility, 0)
	craftingAbilities := make(map[string]CraftingAbility, 0)
	trueAbilities := make(map[string]TrueAbility, 0)

	// allDetails := ""

	// var wg sync.WaitGroup
	// var wgNum int = 5

	// type scrapperChannel struct {
	// }
	// sem := make(chan struct{}, wgNum)

	for _, href := range hrefs {

		// fmt.Println("Navigating to:", href)

		// Create a new tab/context for each navigation
		tabCtx, tabCancel := chromedp.NewContext(ctx)

		hrefArray := strings.Split(href, "/")
		classID := hrefArray[len(hrefArray)-1]
		fmt.Println("Processing class:", classID)

		var detailsHTML string
		var classTier int
		var className string
		var abilitySegments []string
		err := chromedp.Run(tabCtx,
			chromedp.Navigate(href),
			chromedp.WaitVisible(`app-class-details`, chromedp.ByQuery),
			chromedp.Sleep(5*time.Second),
			chromedp.OuterHTML("app-class-details", &detailsHTML),
			chromedp.Evaluate(` Array.from(document.querySelectorAll('.fa-solid.fa-circle.ng-star-inserted')).length `, &classTier),
			chromedp.Evaluate(` Array.from(document.querySelectorAll('mat-expansion-panel')).map(panel => panel.outerHTML) `, &abilitySegments),
			chromedp.Text("h2", &className, chromedp.ByQuery),
		)
		tabCancel()
		if err != nil {
			fmt.Println("Error getting details from", href, ":", err)
			continue
		}

		classInfo := ClassInfo{
			Name: className,
			ID:   classID,
			Tier: classTier,
		}
		classes[classID] = classInfo

		for i, html := range abilitySegments {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				fmt.Println("Error parsing segment:", err)
				continue
			}

			rawName := doc.Find("mat-panel-title span").Text()
			abilityName := strings.Split(rawName, "○ ")[1]
			abilityID := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(abilityName, " ", "_"), "'", ""))
			// fmt.Printf("Ability %d: %s [%s]\n", i+1, abilityName, abilityID)
			var benefits []string
			if i == 0 {
				doc.Find("app-key-ability mat-card-content").Each(func(j int, s *goquery.Selection) {
					s.Find("ul").First().ChildrenFiltered("li").Each(func(j int, li *goquery.Selection) {
						benefits = append(benefits, li.Text())
					})
				})
				if val, ok := keyAbilities[abilityID]; !ok {
					keyAbilities[abilityID] = KeyAbility{
						ID:       abilityID,
						ClassID:  []string{classID},
						Name:     abilityName,
						Benefits: benefits,
					}
				} else {
					keyAbilities[abilityID] = KeyAbility{
						ID:       abilityID,
						ClassID:  append(val.ClassID, classID),
						Name:     abilityName,
						Benefits: benefits,
					}
				}

				// fmt.Printf("Ability %d: %s [%s] with %d benefits\n", i+1, abilityName, abilityID, len(benefits))
				doc.Find("app-true-ability mat-card-title").Each(func(j int, s *goquery.Selection) {
					associatedAbilityName := s.Text()
					if associatedAbilityName == "" {
						associatedAbilityName = "None"
					} else {
						if strings.Contains(associatedAbilityName, "Core Crafting") || strings.Contains(associatedAbilityName, "Expert") || strings.Contains(associatedAbilityName, "Basics") {
							var checkpoints []string
							doc.Find("app-true-ability mat-card-content li strong").Each(func(j int, s *goquery.Selection) {
								checkpoints = append(checkpoints, s.Text())
							})
							// fmt.Printf("  Associated Crafting Ability: %s with %d checkpoints\n", associatedAbilityName, len(checkpoints))

							doc.Find("app-true-ability mat-card-content li").Each(func(j int, s *goquery.Selection) {
								if j == 0 {
									return
								}
								text := s.Text()

								_, after, _ := stringBeforeAfter(text, "Description")

								text = after

								var craftingAbilitiesRaw []CraftingAbilityRaw

								var craftAbilRaw CraftingAbilityRaw

								var newCheckpoints []string
								for _, cp := range checkpoints {
									if strings.Contains(cp, "Keywords") || strings.Contains(cp, "Cost") || strings.Contains(cp, "Description") {
										continue
									} else {
										newCheckpoints = append(newCheckpoints, cp)
									}
								}
								checkpoints = newCheckpoints

								for i, checkpoint := range checkpoints {
									fmt.Println("checkpoint:", checkpoint)
									// fmt.Printf("      Found Other [%s] checkpoint:\n\t %s\n", checkpoint, printNCharacters(text, 50))
									craftingAbilName := checkpoint
									craftingAbilID := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(craftingAbilName), " ", "_"), "'", ""), "’", ""))
									before, after, _ := stringBeforeAfter(text, checkpoint)

									if i != 0 {
										craftAbilRaw.Text = strings.TrimSpace(before)
										// spew.Dump(craftingAbility)
										craftingAbilitiesRaw = append(craftingAbilitiesRaw, craftAbilRaw)
									}
									craftAbilRaw = CraftingAbilityRaw{
										ID:   craftingAbilID,
										Name: strings.TrimSpace(craftingAbilName),
									}
									text = strings.TrimSpace(after)
									if i == len(checkpoints)-1 {
										craftAbilRaw.Text = strings.TrimSpace(after)
										craftingAbilitiesRaw = append(craftingAbilitiesRaw, craftAbilRaw)
									}
								}
								spew.Dump(craftingAbilitiesRaw)

								// var craftingAbilities []CraftingAbility

								var craftingAbility CraftingAbility

								for _, car := range craftingAbilitiesRaw {
									fmt.Println("  Processing Crafting Ability:", car.Name)
									craftingAbility.ID = car.ID
									craftingAbility.Name = car.Name

									text := car.Text

									var cost, keywords bool

									for i, checkpoint := range []string{"Keywords", "Cost", "Description"} {
										if strings.Contains(text, checkpoint) {
											if checkpoint == "Keywords" {
												keywords = true
												_, after, _ := stringBeforeAfter(text, checkpoint)
												text = strings.TrimSpace(after)
											} else if checkpoint == "Cost" {
												cost = true
												before, after, _ := stringBeforeAfter(text, checkpoint)
												craftingAbility.Keywords = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(before), "\u00a0", ""), ":", "")
												text = strings.TrimSpace(after)
											} else if checkpoint == "Description" {
												before, after, _ := stringBeforeAfter(text, checkpoint)
												if cost {
													craftingAbility.Cost = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(before), "\u00a0", ""), ":", "")
												} else if keywords {
													craftingAbility.Keywords = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(before), "\u00a0", ""), ":", "")
												}
												if i == 2 {
													craftingAbility.Description = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(after), "\u00a0", ""), ":", "")
												}
												text = strings.TrimSpace(after)
											}
										}
									}
									// craftingAbilities = append(craftingAbilities, craftingAbility)
									craftingAbilities[craftingAbility.ID] = craftingAbility
									// 			craftingAbilities[craftingAbility.ID] = craftingAbility
								}
							})
						} else {
							abilityName := associatedAbilityName
							abilityID := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(abilityName, " ", "_"), "'", ""), "’", ""))

							var trueAbility TrueAbility
							trueAbility.ID = abilityID
							trueAbility.LevelRequired = levelAbilities[i]
							trueAbility.ClassID = []string{classID}
							trueAbility.Name = abilityName

							// fmt.Printf("  Processing True Ability: %s [%s]\n", abilityName, abilityID)
							doc.Find("app-true-ability mat-card-content li").Each(func(j int, s *goquery.Selection) {
								text := s.Text()

								if strings.Contains(text, "Keywords") {
									keywords := ""
									doc.Find("app-true-ability mat-card-content mat-chip").Each(func(j int, s *goquery.Selection) {
										if j != 0 {
											keywords += ", "
										}
										keywords += s.Text()
									})
									// fmt.Println("    Processing line with Keywords:", text)
									// keywords := strings.Split(text, "Keywords")[1]
									trueAbility.Keywords = keywords
								} else if strings.Contains(text, "Range") {
									rng := strings.Split(text, "Range")[1]
									trueAbility.Range = strings.TrimSpace(rng)
								} else if strings.Contains(text, "Description") {
									desc := strings.Split(text, "Description")[1]
									trueAbility.Description = strings.TrimSpace(desc)
								} else if strings.Contains(text, "Requirement") {
									req := strings.Split(text, "Requirement")[1]
									trueAbility.Requirement = strings.TrimSpace(req)
								} else if strings.Contains(text, "RP cost") {
									rp := strings.Split(text, "RP cost")[1]
									trueAbility.RPcost = strings.TrimSpace(rp)
								} else if strings.Contains(text, "AP cost") {
									ap := strings.Split(text, "AP cost")[1]
									trueAbility.APcost = strings.TrimSpace(ap)
								} else if strings.Contains(text, "MP cost") {
									mp := strings.Split(text, "MP cost")[1]
									trueAbility.MPcost = strings.TrimSpace(mp)
								} else if strings.Contains(text, "Mana cost") {
									mp := strings.Split(text, "Mana cost")[1]
									trueAbility.MPcost = strings.TrimSpace(mp)
								} else if strings.Contains(text, "Other costs") {
									mp := strings.Split(text, "Other costs")[1]
									trueAbility.MPcost = strings.TrimSpace(mp)
								} else {
									fmt.Printf("    Unrecognized line: %s\n", text)
								}
							})

							if val, ok := trueAbilities[abilityID]; ok {
								trueAbility.ClassID = append(val.ClassID, classID)
							} else {
								trueAbility.ClassID = []string{classID}
							}

							trueAbilities[abilityID] = trueAbility
						}
					}
				})
			} else {
				var trueAbility TrueAbility
				trueAbility.ID = abilityID
				trueAbility.ClassID = []string{classID}
				trueAbility.LevelRequired = levelAbilities[i]
				trueAbility.Name = abilityName

				// fmt.Printf("  Processing True Ability: %s [%s]\n", abilityName, abilityID)
				doc.Find("app-true-ability mat-card-content li").Each(func(j int, s *goquery.Selection) {
					text := s.Text()

					if strings.Contains(text, "Keywords") {
						keywords := ""
						doc.Find("app-true-ability mat-card-content mat-chip").Each(func(j int, s *goquery.Selection) {
							if j != 0 {
								keywords += ", "
							}
							keywords += s.Text()
						})
						// fmt.Println("    Processing line with Keywords:", text)
						// keywords := strings.Split(text, "Keywords")[1]
						trueAbility.Keywords = keywords
					} else if strings.Contains(text, "Range") {
						rng := strings.Split(text, "Range")[1]
						trueAbility.Range = strings.TrimSpace(rng)
					} else if strings.Contains(text, "Description") {
						desc := strings.Split(text, "Description")[1]
						trueAbility.Description = strings.TrimSpace(desc)
					} else if strings.Contains(text, "Requirement") {
						req := strings.Split(text, "Requirement")[1]
						trueAbility.Requirement = strings.TrimSpace(req)
					} else if strings.Contains(text, "RP cost") {
						rp := strings.Split(text, "RP cost")[1]
						trueAbility.RPcost = strings.TrimSpace(rp)
					} else if strings.Contains(text, "AP cost") {
						ap := strings.Split(text, "AP cost")[1]
						trueAbility.APcost = strings.TrimSpace(ap)
					} else if strings.Contains(text, "MP cost") {
						mp := strings.Split(text, "MP cost")[1]
						trueAbility.MPcost = strings.TrimSpace(mp)
					} else if strings.Contains(text, "Mana cost") {
						mp := strings.Split(text, "Mana cost")[1]
						trueAbility.MPcost = strings.TrimSpace(mp)
					} else if strings.Contains(text, "Other costs") {
						mp := strings.Split(text, "Other costs")[1]
						trueAbility.MPcost = strings.TrimSpace(mp)
					} else {
						fmt.Printf("    Unrecognized line: %s\n", text)
					}
				})

				if val, ok := trueAbilities[abilityID]; ok {
					trueAbility.ClassID = append(val.ClassID, classID)
				} else {
					trueAbility.ClassID = []string{classID}
				}

				trueAbilities[abilityID] = trueAbility

			}
			// remove in the future
			// break
		}

		// allDetails += detailsHTML + "\n"

		// if k == 20 {
		// 	break
		// }

	}

	// err = os.WriteFile("class_details.html", []byte(allDetails), 0644)
	// if err != nil {
	// 	fmt.Println("Error writing file:", err)
	// 	return
	// }

	// write the classes and keyAbilities maps to JSON files
	classesBytes, err := json.MarshalIndent(classes, " ", "  ")
	if err != nil {
		fmt.Println("Error marshalling classes to JSON:", err)
		return
	}
	err = os.WriteFile("classes.json", classesBytes, 0644)
	if err != nil {
		fmt.Println("Error writing classes.json file:", err)
		return
	}

	keyAbilitiesBytes, err := json.MarshalIndent(keyAbilities, " ", "  ")
	if err != nil {
		fmt.Println("Error marshalling key abilities to JSON:", err)
		return
	}
	err = os.WriteFile("key_abilities.json", keyAbilitiesBytes, 0644)
	if err != nil {
		fmt.Println("Error writing key_abilities.json file:", err)
		return
	}

	craftingAbilitiesBytes, err := json.MarshalIndent(craftingAbilities, " ", "  ")
	if err != nil {
		fmt.Println("Error marshalling crafting abilities to JSON:", err)
		return
	}
	err = os.WriteFile("crafting_abilities.json", craftingAbilitiesBytes, 0644)
	if err != nil {
		fmt.Println("Error writing crafting_abilities.json file:", err)
		return
	}

	trueAbilitiesBytes, err := json.MarshalIndent(trueAbilities, " ", "  ")
	if err != nil {
		fmt.Println("Error marshalling true abilities to JSON:", err)
		return
	}
	err = os.WriteFile("true_abilities.json", trueAbilitiesBytes, 0644)
	if err != nil {
		fmt.Println("Error writing true_abilities.json file:", err)
		return
	}

	fmt.Printf("Wrote %d classes to classes.json\n", len(classes))
	fmt.Printf("Wrote %d key abilities to key_abilities.json\n", len(keyAbilities))
	fmt.Printf("Wrote %d crafting abilities to crafting_abilities.json\n", len(craftingAbilities))
	fmt.Printf("Wrote %d true abilities to true_abilities.json\n", len(trueAbilities))

}

func getRacesAndAbilities() {
	fmt.Println("Getting Races and Abilities...")
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var outerHTML string
	var outerHTML2 string
	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL+"/races"),
		chromedp.WaitVisible(`app-ancestry-card`, chromedp.ByQuery),
		chromedp.Sleep(5*time.Second),
		chromedp.OuterHTML("app-races", &outerHTML),
		chromedp.Click(`app-races .mat-mdc-tab-labels > div[role="tab"]:nth-child(2)`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		chromedp.OuterHTML("app-races", &outerHTML2),
	)
	if err != nil {
		fmt.Println("Error getting hrefs:", err)
		return
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(outerHTML))
	if err != nil {
		fmt.Println("Error parsing segment:", err)
		return
	}

	var hrefs []string

	doc.Find("app-ancestry-card a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			hrefs = append(hrefs, href)
		}
	})

	doc2, err := goquery.NewDocumentFromReader(strings.NewReader(outerHTML2))
	if err != nil {
		fmt.Println("Error parsing segment:", err)
		return
	}

	var hrefs2 []string

	doc2.Find("app-ancestry-card a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			hrefs2 = append(hrefs2, href)
		}
	})

	var races = make(map[string]string, 0)
	var subraces = make(map[string]string, 0)
	var trueAbilities = make(map[string]TrueAbility, 0)

	for _, href := range hrefs {
		// Create a new tab/context for each navigation
		tabCtx, tabCancel := chromedp.NewContext(ctx)

		race := path.Base(href)
		fmt.Println("Processing race:", race)

		var outerHTML string
		var abilitySegments []string
		err := chromedp.Run(tabCtx,
			chromedp.Navigate("https://rpg.angelssword.com"+href),
			chromedp.WaitVisible(`app-primary-details`, chromedp.ByQuery),
			chromedp.Sleep(5*time.Second),
			chromedp.OuterHTML("app-primary-details", &outerHTML),
			chromedp.Evaluate(` Array.from(document.querySelectorAll('mat-expansion-panel')).map(panel => panel.outerHTML) `, &abilitySegments),
		)
		tabCancel()
		if err != nil {
			fmt.Println("Error getting details from", href, ":", err)
			continue
		}

		docx, err := goquery.NewDocumentFromReader(strings.NewReader(outerHTML))
		if err != nil {
			fmt.Println("Error parsing segment:", err)
			continue
		}
		rawRaceName := docx.Find("h2").Text()

		races[race] = rawRaceName

		demon := false
		if race == "demon" {
			demon = true
			fmt.Println("Found Demon Race, extra processing...")
			docx.Find("app-primary-details mat-panel-title").Each(func(j int, s *goquery.Selection) {
				text := s.Text()
				if strings.Contains(text, "-") {
					text = strings.Split(text, " ")[1]
					subraces[strings.ReplaceAll(strings.ToLower(text), "'", "x")] = text
				}
			})
		}

		for _, html := range abilitySegments {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				fmt.Println("Error parsing segment:", err)
				continue
			}

			rawName := doc.Find("mat-panel-title span").Text()
			fmt.Println("checking the text rawName", rawName)

			abilityName := strings.Split(rawName, "○ ")[1]
			abilityID := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(abilityName, " ", "_"), "'", ""))

			var trueAbility TrueAbility
			trueAbility.ID = abilityID
			trueAbility.Name = abilityName
			if demon {
				if strings.Contains(abilityName, "-") {
					abilityNameRaw := strings.ReplaceAll(strings.ToLower(abilityName), "'", "x")
					subrace := strings.Split(abilityNameRaw, " ")[0]
					trueAbility.SubraceID = append(trueAbility.SubraceID, subrace)
				} else {
					trueAbility.RaceID = append(trueAbility.RaceID, race)
				}
			} else {
				trueAbility.RaceID = append(trueAbility.RaceID, race)
			}

			doc.Find("app-true-ability mat-card-content li").Each(func(j int, s *goquery.Selection) {
				text := s.Text()

				if strings.Contains(text, "Keywords") {
					keywords := ""
					doc.Find("app-true-ability mat-card-content mat-chip").Each(func(j int, s *goquery.Selection) {
						if j != 0 {
							keywords += ", "
						}
						keywords += s.Text()
					})
					trueAbility.Keywords = keywords
				} else if strings.Contains(text, "Range") {
					rng := strings.Split(text, "Range")[1]
					trueAbility.Range = strings.TrimSpace(rng)
				} else if strings.Contains(text, "Description") {
					desc := strings.Split(text, "Description")[1]
					trueAbility.Description = strings.TrimSpace(desc)
				} else if strings.Contains(text, "Requirement") {
					req := strings.Split(text, "Requirement")[1]
					trueAbility.Requirement = strings.TrimSpace(req)
				} else if strings.Contains(text, "RP cost") {
					rp := strings.Split(text, "RP cost")[1]
					trueAbility.RPcost = strings.TrimSpace(rp)
				} else if strings.Contains(text, "AP cost") {
					ap := strings.Split(text, "AP cost")[1]
					trueAbility.APcost = strings.TrimSpace(ap)
				} else if strings.Contains(text, "MP cost") {
					mp := strings.Split(text, "MP cost")[1]
					trueAbility.MPcost = strings.TrimSpace(mp)
				} else if strings.Contains(text, "Mana cost") {
					mp := strings.Split(text, "Mana cost")[1]
					trueAbility.MPcost = strings.TrimSpace(mp)
				} else if strings.Contains(text, "Other costs") {
					mp := strings.Split(text, "Other costs")[1]
					trueAbility.MPcost = strings.TrimSpace(mp)
				} else {
					fmt.Printf("    Unrecognized line: %s\n", text)
				}
			})

			trueAbilities[abilityID] = trueAbility
		}

	}

	for _, href := range hrefs2 {

		// Create a new tab/context for each navigation
		tabCtx, tabCancel := chromedp.NewContext(ctx)

		subrace := strings.ReplaceAll(path.Base(href), ";returnTo=secondary", "")
		fmt.Println("Processing sub-race:", subrace)

		var abilitySegments []string
		var outerHTML string
		url := "https://rpg.angelssword.com/game/latest/races/secondary/" + subrace
		err := chromedp.Run(tabCtx,
			chromedp.Navigate(url),
			chromedp.WaitVisible(`app-race-details`, chromedp.ByQuery),
			chromedp.Sleep(5*time.Second),
			chromedp.OuterHTML("app-race-details", &outerHTML),
			chromedp.Evaluate(` Array.from(document.querySelectorAll('mat-expansion-panel')).map(panel => panel.outerHTML) `, &abilitySegments),
		)
		tabCancel()
		if err != nil {
			fmt.Println("Error getting details from", href, ":", err)
			continue
		}

		docx, err := goquery.NewDocumentFromReader(strings.NewReader(outerHTML))
		if err != nil {
			fmt.Println("Error parsing segment:", err)
			continue
		}
		rawSubRaceName := docx.Find("h2").Text()

		subraces[subrace] = rawSubRaceName

		for _, html := range abilitySegments {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				fmt.Println("Error parsing segment:", err)
				continue
			}

			rawName := doc.Find("mat-panel-title span").Text()
			abilityName := strings.Split(rawName, "○ ")[1]
			abilityID := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(abilityName, " ", "_"), "'", ""))
			// fmt.Printf("Ability %d: %s [%s]\n", i+1, abilityName, abilityID)

			var trueAbility TrueAbility
			trueAbility.ID = abilityID
			trueAbility.Name = abilityName
			trueAbility.SubraceID = append(trueAbility.SubraceID, subrace)

			// fmt.Printf("  Processing True Ability: %s [%s]\n", abilityName, abilityID)
			doc.Find("app-true-ability mat-card-content li").Each(func(j int, s *goquery.Selection) {
				text := s.Text()

				if strings.Contains(text, "Keywords") {
					keywords := ""
					doc.Find("app-true-ability mat-card-content mat-chip").Each(func(j int, s *goquery.Selection) {
						if j != 0 {
							keywords += ", "
						}
						keywords += s.Text()
					})
					// fmt.Println("    Processing line with Keywords:", text)
					// keywords := strings.Split(text, "Keywords")[1]
					trueAbility.Keywords = keywords
				} else if strings.Contains(text, "Range") {
					rng := strings.Split(text, "Range")[1]
					trueAbility.Range = strings.TrimSpace(rng)
				} else if strings.Contains(text, "Description") {
					desc := strings.Split(text, "Description")[1]
					trueAbility.Description = strings.TrimSpace(desc)
				} else if strings.Contains(text, "Requirement") {
					req := strings.Split(text, "Requirement")[1]
					trueAbility.Requirement = strings.TrimSpace(req)
				} else if strings.Contains(text, "RP cost") {
					rp := strings.Split(text, "RP cost")[1]
					trueAbility.RPcost = strings.TrimSpace(rp)
				} else if strings.Contains(text, "AP cost") {
					ap := strings.Split(text, "AP cost")[1]
					trueAbility.APcost = strings.TrimSpace(ap)
				} else if strings.Contains(text, "MP cost") {
					mp := strings.Split(text, "MP cost")[1]
					trueAbility.MPcost = strings.TrimSpace(mp)
				} else if strings.Contains(text, "Mana cost") {
					mp := strings.Split(text, "Mana cost")[1]
					trueAbility.MPcost = strings.TrimSpace(mp)
				} else if strings.Contains(text, "Other costs") {
					oc := strings.Split(text, "Other costs")[1]
					trueAbility.Othercost = strings.TrimSpace(oc)
				} else {
					fmt.Printf("    Unrecognized line: %s\n", text)
				}
			})

			trueAbilities[abilityID] = trueAbility

		}

	}

	racesBytes, err := json.MarshalIndent(races, " ", "  ")
	if err != nil {
		fmt.Println("Error marshalling")
		return
	}
	if err = os.WriteFile("races.json", racesBytes, 0644); err != nil {
		fmt.Println("Error writing")
		return
	}

	subracesBytes, err := json.MarshalIndent(subraces, " ", "  ")
	if err != nil {
		fmt.Println("Error marshalling")
		return
	}
	if err = os.WriteFile("subraces.json", subracesBytes, 0644); err != nil {
		fmt.Println("Error writing")
		return
	}

	abilitiesBytes, err := json.MarshalIndent(trueAbilities, " ", "  ")
	if err != nil {
		fmt.Println("Error marshalling true abilities to JSON:", err)
		return
	}
	if err = os.WriteFile("races_true_abilities.json", abilitiesBytes, 0644); err != nil {
		fmt.Println("Error writing true_abilities.json file:", err)
		return
	}

	fmt.Printf("Wrote %d true abilities to class_true_abilities.json\n", len(trueAbilities))
	fmt.Printf("Wrote %d races to races.json\n", len(races))
	fmt.Printf("Wrote %d subraces to subraces.json\n", len(subraces))
}
