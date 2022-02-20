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

func (Decide) Execute(cmd, rest string, c *irc.Client, m *irc.Message) {
	params := strings.Split(rest, " ")

	if cmd == ".decide" {
		var terms []string
		var currTerm strings.Builder
		for _, word := range params {
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
			c.WriteMessage(NewRes(m, "Usage: .decide proposition 1 [ or proposition 2 [ or proposition n ... ] ]"))
			return
		} else if len(terms) < 2 {
			c.WriteMessage(NewRes(m, []string{"Yes.", "No."}[rand.Intn(2)]))
			return
		} else {
			c.WriteMessage(NewRes(m, terms[rand.Intn(len(terms))]))
			return
		}
	} else if cmd == ".dice" || cmd == ".roll" {
		dice := params[0]
		if dice == "" {
			c.WriteMessage(NewRes(m, "usage: .dice NNdXX - where NN is 1-36 and XX is 2-64"))
			return
		}
		if len(dice) > 5 {
			c.WriteMessage(NewRes(m, "Invalid dice specification: too big"))
			return
		}

		spec := strings.SplitN(dice, "d", 2)
		if len(spec) != 2 {
			c.WriteMessage(NewRes(m, "Invalid dice specification: no separating 'd'"))
			return
		}

		numDie, err := strconv.Atoi(spec[0])
		if err != nil || numDie < 1 || numDie > 36 {
			c.WriteMessage(NewRes(m, fmt.Sprintf("Invalid dice count: %s is not a number or is not between 1-36", spec[0])))
			return
		}
		numDieFaces, err := strconv.Atoi(spec[1])
		if err != nil || numDieFaces < 2 || numDieFaces > 64 {
			c.WriteMessage(NewRes(m, fmt.Sprintf("Invalid dice face count: %s is not a number or is not between 2-64", spec[0])))
			return
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
		c.WriteMessage(NewRes(m, result.String()))
		return
	}

	panic("Unreachable!")
}
