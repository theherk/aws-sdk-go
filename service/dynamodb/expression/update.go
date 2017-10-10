package expression

import (
	"fmt"
	"sort"
	"strings"
)

// operationMode specifies the types of update operations that the
// updateBuilder is going to represent. The const is in a string to use the
// const value as a map key and as a string when creating the formatted
// expression for the exprNodes.
type operationMode string

const (
	setOperation    operationMode = "SET"
	removeOperation               = "REMOVE"
	addOperation                  = "ADD"
	deleteOperation               = "DELETE"
)

var updateOps = []operationMode{
	addOperation,
	deleteOperation,
	removeOperation,
	setOperation,
}

// Implementing the Sort interface
type modeList []operationMode

func (ml modeList) Len() int {
	return len(ml)
}

func (ml modeList) Less(i, j int) bool {
	return string(ml[i]) < string(ml[j])
}

func (ml modeList) Swap(i, j int) {
	ml[i], ml[j] = ml[j], ml[i]
}

// UpdateBuilder represents Update Expressions in DynamoDB. UpdateBuilders
// are the building blocks of the Builder struct. Note that there are different
// update operations in DynamoDB and an UpdateBuilder can represent multiple
// update operations.
// More Information at: http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.UpdateExpressions.html
type UpdateBuilder struct {
	operationList map[operationMode][]operationBuilder
}

// operationBuilder represents specific update actions (SET, REMOVE, ADD,
// DELETE). The mode specifies what type of update action the
// operationBuilder represents.
type operationBuilder struct {
	name  NameBuilder
	value OperandBuilder
	mode  operationMode
}

// buildOperation builds an exprNode from an operationBuilder. buildOperation
// is called recursively by buildTree in order to create a tree structure
// of exprNodes representing the parent/child relationships between
// UpdateBuilders and operationBuilders.
func (ob operationBuilder) buildOperation() (exprNode, error) {
	pathChild, err := ob.name.BuildOperand()
	if err != nil {
		return exprNode{}, err
	}

	node := exprNode{
		children: []exprNode{pathChild.exprNode},
		fmtExpr:  "$c",
	}

	if ob.mode == removeOperation {
		return node, nil
	}

	valueChild, err := ob.value.BuildOperand()
	if err != nil {
		return exprNode{}, err
	}
	node.children = append(node.children, valueChild.exprNode)

	switch ob.mode {
	case setOperation:
		node.fmtExpr += " = $c"
	case addOperation, deleteOperation:
		node.fmtExpr += " $c"
	default:
		return exprNode{}, fmt.Errorf("build update error: build operation error: unsupported mode: %v", ob.mode)
	}

	return node, nil
}

// Delete returns an UpdateBuilder representing one Delete operation for
// DynamoDB Update Expressions. The argument name should specify the item
// attribute and the argument value should specify the value to be deleted. The
// resulting UpdateBuilder can be used as an argument to the WithUpdate() method
// for the Builder struct.
//
// Example:
//
//     // update represents the delete operation to delete the string value
//     // "subsetToDelete" from the item attribute "pathToList"
//     update := expression.Delete(expression.Name("pathToList"), expression.Value("subsetToDelete"))
//
//     // Adding more update methods
//     anotherUpdate := update.Remove(expression.Name("someName"))
//     // Creating a Builder
//     builder := Update(update)
//
// Expression Equivalent:
//
//     expression.Delete(expression.Name("pathToList"), expression.Value("subsetToDelete"))
//     // let :del be an ExpressionAttributeValue representing the value
//     // "subsetToDelete"
//     "DELETE pathToList :del"
func Delete(name NameBuilder, value ValueBuilder) UpdateBuilder {
	emptyUpdateBuilder := UpdateBuilder{}
	return emptyUpdateBuilder.Delete(name, value)
}

// Delete adds a Delete operation to the argument UpdateBuilder. The
// argument name should specify the item attribute and the argument value should
// specify the value to be deleted. The resulting UpdateBuilder can be used as
// an argument to the WithUpdate() method for the Builder struct.
//
// Example:
//
//     // Let update represent an already existing update expression. Delete()
//     // adds the operation to delete the value "subsetToDelete" from the item
//     // attribute "pathToList"
//     update := update.Delete(expression.Name("pathToList"), expression.Value("subsetToDelete"))
//
//     // Adding more update methods
//     anotherUpdate := update.Remove(expression.Name("someName"))
//     // Creating a Builder
//     builder := Update(update)
//
// Expression Equivalent:
//
//     Delete(expression.Name("pathToList"), expression.Value("subsetToDelete"))
//     // let :del be an ExpressionAttributeValue representing the value
//     // "subsetToDelete"
//     "DELETE pathToList :del"
func (ub UpdateBuilder) Delete(name NameBuilder, value ValueBuilder) UpdateBuilder {
	if ub.operationList == nil {
		ub.operationList = map[operationMode][]operationBuilder{}
	}
	ub.operationList[deleteOperation] = append(ub.operationList[deleteOperation], operationBuilder{
		name:  name,
		value: value,
		mode:  deleteOperation,
	})
	return ub
}

// Add returns an UpdateBuilder representing the Add operation for DynamoDB
// Update Expressions. The argument name should specify the item attribute and
// the argument value should specify the value to be added. The resulting
// UpdateBuilder can be used as an argument to the WithUpdate() method for the
// Builder struct.
//
// Example:
//
//     // update represents the add operation to add the value 5 to the item
//     // attribute "aPath"
//     update := expression.Add(expression.Name("aPath"), expression.Value(5))
//
//     // Adding more update methods
//     anotherUpdate := update.Remove(expression.Name("someName"))
//     // Creating a Builder
//     builder := Update(update)
//
// Expression Equivalent:
//
//     expression.Add(expression.Name("aPath"), expression.Value(5))
//     // Let :five be an ExpressionAttributeValue representing the value 5
//     "ADD aPath :5"
func Add(name NameBuilder, value ValueBuilder) UpdateBuilder {
	emptyUpdateBuilder := UpdateBuilder{}
	return emptyUpdateBuilder.Add(name, value)
}

// Add adds an Add operation to the argument UpdateBuilder. The argument
// name should specify the item attribute and the argument value should specify
// the value to be added. The resulting UpdateBuilder can be used as an argument
// to the WithUpdate() method for the Builder struct.
//
// Example:
//
//     // Let update represent an already existing update expression. Add() adds
//     // the operation to add the value 5 to the item attribute "aPath"
//     update := update.Add(expression.Name("aPath"), expression.Value(5))
//
//     // Adding more update methods
//     anotherUpdate := update.Remove(expression.Name("someName"))
//     // Creating a Builder
//     builder := Update(update)
//
// Expression Equivalent:
//
//     Add(expression.Name("aPath"), expression.Value(5))
//     // Let :five be an ExpressionAttributeValue representing the value 5
//     "ADD aPath :5"
func (ub UpdateBuilder) Add(name NameBuilder, value ValueBuilder) UpdateBuilder {
	if ub.operationList == nil {
		ub.operationList = map[operationMode][]operationBuilder{}
	}
	ub.operationList[addOperation] = append(ub.operationList[addOperation], operationBuilder{
		name:  name,
		value: value,
		mode:  addOperation,
	})
	return ub
}

// Remove returns an UpdateBuilder representing the Remove operation for
// DynamoDB Update Expressions. The argument name should specify the item
// attribute to delete. The resulting UpdateBuilder can be used as an argument
// to the WithUpdate() method for the Builder struct.
//
// Example:
//
//     // update represents the remove operation to remove the item attribute
//     // "itemToRemove"
//     update := expression.Remove(expression.Name("itemToRemove"))
//
//     // Adding more update methods
//     anotherUpdate := update.Remove(expression.Name("someName"))
//     // Creating a Builder
//     builder := Update(update)
//
// Expression Equivalent:
//
//     expression.Remove(expression.Name("itemToRemove"))
//     "REMOVE itemToRemove"
func Remove(name NameBuilder) UpdateBuilder {
	emptyUpdateBuilder := UpdateBuilder{}
	return emptyUpdateBuilder.Remove(name)
}

// Remove adds a Remove operation to the argument UpdateBuilder. The
// argument name should specify the item attribute to delete. The resulting
// UpdateBuilder can be used as an argument to the WithUpdate() method for the
// Builder struct.
//
// Example:
//
//     // Let update represent an already existing update expression. Remove()
//     // adds the operation to remove the item attribute "itemToRemove"
//     update := update.Remove(expression.Name("itemToRemove"))
//
//     // Adding more update methods
//     anotherUpdate := update.Remove(expression.Name("someName"))
//     // Creating a Builder
//     builder := Update(update)
//
// Expression Equivalent:
//
//     Remove(expression.Name("itemToRemove"))
//     "REMOVE itemToRemove"
func (ub UpdateBuilder) Remove(name NameBuilder) UpdateBuilder {
	if ub.operationList == nil {
		ub.operationList = map[operationMode][]operationBuilder{}
	}
	ub.operationList[removeOperation] = append(ub.operationList[removeOperation], operationBuilder{
		name: name,
		mode: removeOperation,
	})
	return ub
}

// Set returns an UpdateBuilder representing the Set operation for DynamoDB
// Update Expressions. The argument name should specify the item attribute to
// modify. The argument OperandBuilder should specify the value to modify the
// the item attribute to. The resulting UpdateBuilder can be used as an argument
// to the WithUpdate() method for the Builder struct.
//
// Example:
//
//     // update represents the set operation to set the item attribute
//     // "itemToSet" to the value "setValue" if the item attribute does not
//     // exist yet. (conditional write)
//     update := expression.Set(expression.Name("itemToSet"), expression.IfNotExists(expression.Name("itemToSet"), expression.Value("setValue")))
//
//     // Adding more update methods
//     anotherUpdate := update.Remove(expression.Name("someName"))
//     // Creating a Builder
//     builder := Update(update)
//
// Expression Equivalent:
//
//     expression.Set(expression.Name("itemToSet"), expression.IfNotExists(expression.Name("itemToSet"), expression.Value("setValue")))
//     // Let :val be an ExpressionAttributeValue representing the value
//     // "setValue"
//     "SET itemToSet = :val"
func Set(name NameBuilder, operandBuilder OperandBuilder) UpdateBuilder {
	emptyUpdateBuilder := UpdateBuilder{}
	return emptyUpdateBuilder.Set(name, operandBuilder)
}

// Set adds a Set operation to the argument UpdateBuilder. The argument name
// should specify the item attribute to modify. The argument OperandBuilder
// should specify the value to modify the the item attribute to. The resulting
// UpdateBuilder can be used as an argument to the WithUpdate() method for the
// Builder struct.
//
// Example:
//
//     // Let update represent an already existing update expression. Set() adds
//     // the operation to to set the item attribute "itemToSet" to the value
//     // "setValue" if the item attribute does not exist yet. (conditional
//     // write)
//     update := update.Set(expression.Name("itemToSet"), expression.IfNotExists(expression.Name("itemToSet"), expression.Value("setValue")))
//
//     // Adding more update methods
//     anotherUpdate := update.Remove(expression.Name("someName"))
//     // Creating a Builder
//     builder := Update(update)
//
// Expression Equivalent:
//
//     Set(expression.Name("itemToSet"), expression.IfNotExists(expression.Name("itemToSet"), expression.Value("setValue")))
//     // Let :val be an ExpressionAttributeValue representing the value
//     // "setValue"
//     "SET itemToSet = :val"
func (ub UpdateBuilder) Set(name NameBuilder, operandBuilder OperandBuilder) UpdateBuilder {
	if ub.operationList == nil {
		ub.operationList = map[operationMode][]operationBuilder{}
	}
	ub.operationList[setOperation] = append(ub.operationList[setOperation], operationBuilder{
		name:  name,
		value: operandBuilder,
		mode:  setOperation,
	})
	return ub
}

// buildTree builds a tree structure of exprNodes based on the tree
// structure of the input UpdateBuilder's child UpdateBuilders/Operands.
// buildTree() satisfies the TreeBuilder interface so ProjectionBuilder can be a
// part of Expression struct.
func (ub UpdateBuilder) buildTree() (exprNode, error) {
	if ub.operationList == nil {
		return exprNode{}, newUnsetParameterError("buildTree", "UpdateBuilder")
	}
	ret := exprNode{
		children: []exprNode{},
	}

	modes := modeList{}

	for mode := range ub.operationList {
		modes = append(modes, mode)
	}

	sort.Sort(modes)

	for _, key := range modes {
		ret.fmtExpr += string(key) + " $c\n"

		childNode, err := buildChildNodes(ub.operationList[key])
		if err != nil {
			return exprNode{}, err
		}

		ret.children = append(ret.children, childNode)
	}

	return ret, nil
}

// buildChildNodes creates the list of the child exprNodes.
func buildChildNodes(operationBuilderList []operationBuilder) (exprNode, error) {
	if len(operationBuilderList) == 0 {
		return exprNode{}, fmt.Errorf("buildChildNodes error: operationBuilder list is empty")
	}

	node := exprNode{
		children: make([]exprNode, 0, len(operationBuilderList)),
		fmtExpr:  "$c" + strings.Repeat(", $c", len(operationBuilderList)-1),
	}

	for _, val := range operationBuilderList {
		valNode, err := val.buildOperation()
		if err != nil {
			return exprNode{}, err
		}
		node.children = append(node.children, valNode)
	}

	return node, nil
}
// parseExprStr converts an expression string to an exprNode.
// Since each expression type has unique characteristics for how the are
// constructed, each builder struct must have a parseExprStr method
// defined.
func (ub *UpdateBuilder) parseExprStr(input dynamodb.UpdateItemInput) {
	opExprsMap := map[operationMode][]string{}
	remaining := *input.UpdateExpression
	for len(remaining) > 0 {
		opMode, opExprs, remainder := firstOp(remaining)
		opExprsMap[opMode] = opExprs
		remaining = remainder
	}
	for k, v := range opExprsMap {
		ub.operationList[k] = []operationBuilder{}
		for _, expr := range v {
			ub.operationList[k] = append(ub.operationList[k], opBuilderFromExpr(
				k, expr, input.ExpressionAttributeNames, input.ExpressionAttributeValues),
			)
		}
	}
}

// firstOp takes the portion of an expression string following the first
// operation keyword and returns the substring that leads to the next
// keyword if found. It returns the first operation mode found, its
// actions, and the string following the first operation.
// Its purpose is to help break update expressions into there parts.
// As an example, when parsing the expression "SET x = y, REMOVE z",
// After extracting "SET", you must find the remained that applies to
// the set operation.
// To do so, we split at each of the remaining keywords, leaving just
// the set operation's actions.
func firstOp(expr string) (operationMode, []string, string) {
	expr = strings.Replace(expr, "\n", " ", -1)
	var opMode operationMode
	var opCompoundExpr, remainder string
	candidate := expr
	for _, v := range updateOps {
		parts := strings.SplitN(candidate, string(v), 2)
		if len(parts[0]) == 0 {
			opMode = v
			remainder = parts[1]
			break
		}
		candidate = parts[0]
		if len(parts) > 1 {
			remainder = string(v) + parts[1]
		}
	}
	opCompoundExpr = remainder
	for _, v := range updateOps {
		if v != opMode {
			opCompoundExpr = strings.SplitN(opCompoundExpr, string(v), 2)[0]
		}
	}
	remainder = strings.Trim(remainder[len(opCompoundExpr):], ", ")
	opCompoundExpr = strings.Trim(opCompoundExpr, ", ")
	opExprs := strings.Split(opCompoundExpr, ",")
	for i, v := range opExprs {
		opExprs[i] = strings.Trim(v, " ")
	}
	return opMode, opExprs, remainder
}
