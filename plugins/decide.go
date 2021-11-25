package plugins

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"gopkg.in/irc.v3"
)

type Decide struct{}

func init() {
	Register(Decide{})
}

func (Decide) Triggers() []string {
	return []string{".decide", ".dice", ".roll"}
}

func (Decide) Execute(m *irc.Message) (string, error) {
	params := strings.Split(m.Trailing(), " ")

	trigger := params[0]
	if trigger == ".decide" {
		var terms []string
		var currTerm strings.Builder
		for _, word := range params[1:] {
			if word == "" {
				continue
			}

			if word == "or" && currTerm.Len() != 0 {
				terms = append(terms, currTerm.String())
				currTerm.Reset()
			} else {
				currTerm.WriteString(word)
				currTerm.WriteByte(' ')
			}
		}
		if currTerm.Len() > 0 {
			terms = append(terms, currTerm.String())
		}

		if len(terms) < 1 {
			return "usage: .decide proposition 1 [ or proposition 2 [ or proposition n ... ] ]", nil
		} else if len(terms) < 2 {
			return []string{"Yes.", "No."}[rand.Intn(2)], nil
		} else {
			return terms[rand.Intn(len(terms))], nil
		}
	} else if trigger == ".dice" || trigger == ".roll" {
		if len(params) != 2 {
			return "usage: .dice NNdXX - where NN is 1-36 and XX is 2-64", nil
		}
		dice := params[1]
		if len(dice) > 5 {
			return "Invalid dice specification: too big", nil
		}

		spec := strings.SplitN(dice, "d", 2)
		if len(spec) != 2 {
			return "Invalid dice specification: no separating 'd'", nil
		}

		numDie, err := strconv.Atoi(spec[0])
		if err != nil || numDie < 1 || numDie > 36 {
			return fmt.Sprintf("Invalid dice count: %s is not a number or is not between 1-36", spec[0]), nil
		}
		numDieFaces, err := strconv.Atoi(spec[1])
		if err != nil || numDieFaces < 2 || numDieFaces > 64 {
			return fmt.Sprintf("Invalid dice face count: %s is not a number or is not between 2-64", spec[0]), nil
		}

		var result strings.Builder
		sum := 0
		for i := 0; i < numDie; i++ {
			r := rand.Intn(numDieFaces) + 1
			sum += r
			result.WriteString(strconv.Itoa(r))
			result.WriteByte(' ')
		}
		result.WriteByte('=')
		result.WriteByte(' ')
		result.WriteString(strconv.Itoa(sum))
		return result.String(), nil
	}

	panic("Unreachable!")
}
