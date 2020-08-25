package main

import (
	"errors"
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"strings"
)

const PreKey = "req_"

var (
	ErrNoExpr   = errors.New("cel: no expression")
	ErrParsing  = errors.New("cel: error parsing the expression")
	ErrChecking = errors.New("cel: error checking the expression and its param definition")
)

func defaultDeclarations() cel.EnvOption {
	return cel.Declarations(
		decls.NewVar(PreKey+"method", decls.String),
		decls.NewVar(PreKey+"path", decls.String),
		decls.NewVar(PreKey+"params", decls.NewMapType(decls.String, decls.String)),
		decls.NewVar(PreKey+"headers", decls.NewMapType(decls.String, decls.NewListType(decls.String))),
		decls.NewVar(PreKey+"querystring", decls.NewMapType(decls.String, decls.NewListType(decls.String))),
		decls.NewVar(PreKey+"postdata", decls.String),
	)
}

type InterpretableDefinition struct {
	CheckExpression string
	ModExpression   string
}

type ExParser struct {
	extractor func(InterpretableDefinition) string
}

func (p ExParser) Parse(definition InterpretableDefinition) (cel.Program, error) {
	expr := p.extractor(definition)
	if expr == "" {
		return nil, ErrNoExpr
	}
	env, err := cel.NewEnv(defaultDeclarations())
	if err != nil {
		Log.Error(err)
		return nil, err
	}

	ast, iss := env.Parse(p.extractor(definition))
	if iss != nil && iss.Err() != nil {
		Log.Error(iss.Err())
		return nil, ErrParsing
	}
	c, iss := env.Check(ast)
	if iss != nil && iss.Err() != nil {
		Log.Error(iss.Err())
		return nil, ErrChecking
	}
	return env.Program(c)
}

func (p ExParser) parseByKey(definitions []InterpretableDefinition, key string) ([]cel.Program, error) {
	var res []cel.Program
	for _, def := range definitions {
		if !strings.Contains(p.extractor(def), key) {
			continue
		}
		v, err := p.Parse(def)
		if err == ErrNoExpr {
			continue
		}

		if err != nil {
			return res, err
		}
		res = append(res, v)
	}
	return res, nil
}

func (p ExParser) ParsePre(denitions []InterpretableDefinition) ([]cel.Program, error) {
	return p.parseByKey(denitions, PreKey)
}

func newReqActivation(r *Request) map[string]interface{} {
	return map[string]interface{}{
		PreKey + "method":   r.Method,
		PreKey + "headers":  r.Headers,
		PreKey + "postdata": r.Postdata,
	}
}

func evalChecks(args map[string]interface{}, ps []cel.Program) (bool, error) {
	for _, eval := range ps {
		res, _, err := eval.Eval(args)
		if err != nil {
			Log.Error(err)
			return false, err
		}
		//if !(res.Value().(bool)) {
		//	return false, nil
		//}
		if v, ok := res.Value().(bool); !ok || !v {
			return false, fmt.Errorf("CEL: request aborted by %+v", eval)
		}
	}
	return true, nil
}

func extractCheckExpr(i InterpretableDefinition) string {
	return i.CheckExpression
}

func NewCheckExpressionParser() ExParser {
	return ExParser{
		extractor: extractCheckExpr,
	}
}

func Check(defs []InterpretableDefinition, r *Request) (bool, error) {
	p := NewCheckExpressionParser()
	preEvaluators, err := p.ParsePre(defs)
	if err != nil {
		Log.Error(err)
	}
	result, err := evalChecks(newReqActivation(r), preEvaluators)
	return result, err
}
