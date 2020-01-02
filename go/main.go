package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Container struct {
	Operation string `json:"operation"`
	Left interface{} `json:"left"`
	Right interface{} `json:"right"`
}

func (c Container) isValid() bool {
	return c.Operation != "" && c.Left != nil && c.Right != nil
}


type Result struct {
	Result float64 `json:"result"`
}

func doOperation(op string, left float64, right float64) (float64, error) {
	switch op {
	case "+":
		return left + right, nil
	case "-":
		return left - right, nil
	case "*":
		return left * right, nil
	case "/":
		if right == 0.0 {
			return .0, errors.New("zero division error")
		}
		return left / right, nil
	default:
		return .0, errors.New("unsupported operation")
	}
}

func collapseTree(tree Container) (float64, error) {
	var err error = nil

	if !tree.isValid() {
		return .0, errors.New("wrong data")
	}

	// internal node
	var left float64
	switch tree.Left.(type) {
	case float64:
		left = tree.Left.(float64)
	case map[string]interface{}:
		subtree, err := parseContainer(tree.Left.(map[string]interface{}))
		if err == nil {
			left, err = collapseTree(subtree)
		}
	default:
		err = errors.New("unexpected data type")
	}

	var right float64
	switch tree.Right.(type) {
	case float64:
		right = tree.Right.(float64)
	case map[string]interface{}:
		subtree, err := parseContainer(tree.Right.(map[string]interface{}))
		if err == nil {
			right, err = collapseTree(subtree)
		}
	default:
		err = errors.New("unexpected data type")
	}

	result, err := doOperation(tree.Operation, left, right)
	if err != nil {
		return 0.0, err
	}

	return result, nil
}

func returnError(w http.ResponseWriter, e error) {
	w.WriteHeader(400)
	_, _ = w.Write([]byte(e.Error()))
}

func parseContainer(candidate map[string]interface{}) (Container, error){
	newC := Container{}
	op, ok := candidate["operation"].(string)
	if !ok {
		return newC, errors.New("unexpected data type")
	} else {
		newC.Operation = op
	}
	newC.Left, ok = candidate["left"]
	newC.Right, ok = candidate["right"]
	if !ok {
		return newC, errors.New("unexpected data type")
	}
	return newC, nil
}

func calcHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.Method != "POST" {
		err = errors.New("use POST, Luke")
		returnError(w, err)
		return
	}

	// parse data
	var tree interface{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&tree)
	if err != nil {
		returnError(w, err)
		return
	}

	var answer float64
	switch tree.(type) {
	case float64:
		answer = tree.(float64)
	case map[string]interface{}:
		newTree, err := parseContainer(tree.(map[string]interface{}))
		if err == nil {
			answer, err = collapseTree(newTree)
		}
	default:
		err = errors.New("unexpected data type")
	}

	if err != nil {
		returnError(w, err)
		return
	}

	result := Result{answer}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}


func main() {
	http.HandleFunc("/calc", calcHandler)
	_ = http.ListenAndServe("0.0.0.0:8330", nil)
}