package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	result float64
	expr   string
)

func evalExpression(expr string) (float64, error) {
	expr = strings.ReplaceAll(expr, " ", "")
	tokens := []string{}
	var num strings.Builder

	for i, ch := range expr {
		if ch >= '0' && ch <= '9' || ch == '.' {
			num.WriteRune(ch)
		} else {
			if num.Len() > 0 {
				tokens = append(tokens, num.String())
				num.Reset()
			}
			if ch == '-' && (i == 0 || strings.ContainsAny(tokens[len(tokens)-1], "+-*÷")) {
				num.WriteRune(ch)
			} else {
				tokens = append(tokens, string(ch))
			}
		}
	}

	if num.Len() > 0 {
		tokens = append(tokens, num.String())
	}

	stack := []float64{}
	operator := ""

	for _, token := range tokens {
		if token == "+" || token == "-" || token == "*" || token == "÷" {
			operator = token
		} else {
			num, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid number: %s", token)
			}
			stack = append(stack, num)

			if operator != "" && len(stack) >= 2 {
				right := stack[len(stack)-1]
				left := stack[len(stack)-2]
				stack = stack[:len(stack)-2]

				var res float64
				switch operator {
				case "+":
					res = left + right
				case "-":
					res = left - right
				case "*":
					res = left * right
				case "÷":
					if right == 0 {
						return 0, fmt.Errorf("division by zero")
					}
					res = left / right
				}
				stack = append(stack, res)
				operator = ""
			}
		}
	}

	if len(stack) != 1 {
		return 0, fmt.Errorf("invalid expression")
	}
	return stack[0], nil
}

func calculatorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		action := r.FormValue("action")

		if action == "C" {
			expr = ""
			result = 0
		} else if action == "X" {
			if len(expr) > 0 {
				expr = expr[:len(expr)-1]
			}
		} else if action == "=" {
			if expr == "" {
				expr = "Error"
			} else {
				res, err := evalExpression(expr)
				if err == nil {
					result = res
					if res == float64(int(res)) {
						expr = fmt.Sprintf("%d", int(res))
					} else {
						expr = strconv.FormatFloat(res, 'f', -1, 64)
					}
				} else {
					expr = "Error"
				}
			}
		} else {
			if expr == "Error" {
				expr = action
			} else {
				lastChar := ""
				if len(expr) > 0 {
					lastChar = string(expr[len(expr)-1])
				}

				if (lastChar == "+" || lastChar == "-" || lastChar == "*" || lastChar == "÷") &&
					(action == "+" || action == "-" || action == "*" || action == "÷") {
					expr = expr[:len(expr)-1]
				}

				if (lastChar == "÷" || lastChar == "+" || lastChar == "-" || lastChar == "*") && action == "-" {
					expr += action
				} else {
					expr += action
				}
			}
		}
	}

	tmplPath := filepath.Join("tmpl", "calculator.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, struct{ Result string }{Result: expr})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/", calculatorHandler)
	fmt.Println("Starting server on :8147")
	if err := http.ListenAndServe(":8147", nil); err != nil {
		log.Fatal(err)
	}
}
