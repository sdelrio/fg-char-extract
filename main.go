package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

func main() {
	filename := "db.xml"
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	if err := run(filename); err != nil {
		slog.Error("Execution failed", "error", err)
		os.Exit(1)
	}
}

type Character struct {
	ID             string
	Level          string
	Tokens         []xml.Token
	ProfBonus      int
	AbilityBonuses map[string]int
}

func run(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("opening %s: %w", filename, err)
	}
	defer f.Close()

	decoder := xml.NewDecoder(f)

	// Regex for ignored tags
	ignoredRegex := regexp.MustCompile(`^(public|holder.*)$`)

	var (
		inCharsheet bool
		currentChar *Character
		xmlDepth    int
		p1, p2, p3  string
	)

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("decoding token: %w", err)
		}

		// Ensure we work with a copy to avoid internal buffer reuse issues
		tok := xml.CopyToken(t)

		switch token := tok.(type) {
		case xml.StartElement:
			if !inCharsheet && (token.Name.Local == "charsheet" || token.Name.Local == "charsheets") {
				inCharsheet = true
				continue
			}

			if inCharsheet {
				if currentChar == nil {
					if ignoredRegex.MatchString(token.Name.Local) {
						if err := skipElement(decoder, token); err != nil {
							return err
						}
						continue
					}
					currentChar = &Character{
						ID:             token.Name.Local,
						Tokens:         []xml.Token{tok},
						AbilityBonuses: make(map[string]int),
					}
					xmlDepth = 1
				} else {
					if ignoredRegex.MatchString(token.Name.Local) {
						if err := skipElement(decoder, token); err != nil {
							return err
						}
						continue
					}

					xmlDepth++
					currentChar.Tokens = append(currentChar.Tokens, tok)
					switch xmlDepth {
					case 2:
						p1 = token.Name.Local
					case 3:
						p2 = token.Name.Local
					case 4:
						p3 = token.Name.Local
					}
				}
			}

		case xml.EndElement:
			if inCharsheet {
				if currentChar != nil {
					currentChar.Tokens = append(currentChar.Tokens, tok)
					if xmlDepth == 1 && token.Name.Local == currentChar.ID {
						// End of character
						if err := writeCharacter(currentChar); err != nil {
							return fmt.Errorf("writing character %s: %w", currentChar.ID, err)
						}
						currentChar = nil
						xmlDepth = 0
					} else {
						xmlDepth--
					}
				} else if token.Name.Local == "charsheet" || token.Name.Local == "charsheets" {
					inCharsheet = false
				}
			}

		case xml.CharData:
			if inCharsheet && currentChar != nil {
				currentChar.Tokens = append(currentChar.Tokens, tok)
				cdStr := string(token)
				if xmlDepth == 2 {
					if p1 == "level" {
						currentChar.Level = cdStr
					} else if p1 == "profbonus" {
						fmt.Sscanf(cdStr, "%d", &currentChar.ProfBonus)
					}
				} else if xmlDepth == 4 {
					if p1 == "abilities" && p3 == "bonus" {
						var bonus int
						fmt.Sscanf(cdStr, "%d", &bonus)
						currentChar.AbilityBonuses[p2] = bonus
					}
				}
			}

		default:
			if inCharsheet && currentChar != nil {
				currentChar.Tokens = append(currentChar.Tokens, tok)
			}
		}
	}

	return nil
}

func skipElement(d *xml.Decoder, start xml.StartElement) error {
	depth := 1
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}
		switch t.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
			if depth == 0 {
				return nil
			}
		}
	}
}

func writeCharacter(c *Character) error {
	level := strings.TrimSpace(c.Level)
	if level == "" {
		level = "0"
	}
	filename := fmt.Sprintf("character_%s_%s.xml", c.ID, level)

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"); err != nil {
		return err
	}
	if _, err := f.WriteString("<root version=\"3.1\" release=\"7|CoreRPG:3\">\n"); err != nil {
		return err
	}
	if _, err := f.WriteString("\t<character>\n"); err != nil {
		return err
	}

	enc := xml.NewEncoder(f)

	var (
		inSkillList      bool
		currentSkillProf int
		currentSkillStat string
		lastStartTag     string
	)

	if len(c.Tokens) > 2 {
		for _, t := range c.Tokens[1 : len(c.Tokens)-1] {
			switch token := t.(type) {
			case xml.StartElement:
				lastStartTag = token.Name.Local
				if token.Name.Local == "skilllist" {
					inSkillList = true
				}
				if err := enc.EncodeToken(t); err != nil {
					return err
				}
			case xml.EndElement:
				if err := enc.EncodeToken(t); err != nil {
					return err
				}
				if inSkillList && token.Name.Local == "stat" {
					abilityBonus := c.AbilityBonuses[currentSkillStat]
					total := abilityBonus + (currentSkillProf * c.ProfBonus)
					if err := enc.Flush(); err != nil {
						return err
					}
					if _, err := f.WriteString(fmt.Sprintf("\n\t\t\t\t\t<total type=\"number\">%d</total>", total)); err != nil {
						return err
					}
				}
				if token.Name.Local == "skilllist" {
					inSkillList = false
				}
			case xml.CharData:
				if inSkillList {
					if lastStartTag == "prof" {
						fmt.Sscanf(string(token), "%d", &currentSkillProf)
					} else if lastStartTag == "stat" {
						currentSkillStat = strings.TrimSpace(string(token))
					}
				}
				if err := enc.Flush(); err != nil {
					return err
				}
				s := string(token)
				s = strings.ReplaceAll(s, "\r", "\n")
				s = strings.ReplaceAll(s, "&", "&amp;")
				s = strings.ReplaceAll(s, "<", "&lt;")
				s = strings.ReplaceAll(s, ">", "&gt;")
				if _, err := f.WriteString(s); err != nil {
					return err
				}
			default:
				if err := enc.EncodeToken(t); err != nil {
					return err
				}
			}
		}
	}

	enc.Flush()
	if _, err := f.WriteString("\n\t</character>\n</root>"); err != nil {
		return err
	}
	slog.Info("Character extracted successfully", "filename", filename)
	return nil
}
