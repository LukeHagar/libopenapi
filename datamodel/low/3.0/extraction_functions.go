package v3

import (
	"github.com/pb33f/libopenapi/datamodel/low"
	"github.com/pb33f/libopenapi/utils"
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
	"sync"
)

func FindItemInCollection[T any](item string, collection map[low.KeyReference[string]]map[low.KeyReference[string]]low.ValueReference[T]) *low.ValueReference[T] {
	for _, c := range collection {
		for n, o := range c {
			if n.Value == item {
				return &o
			}
		}
	}
	return nil
}

func FindItemInMap[T any](item string, collection map[low.KeyReference[string]]low.ValueReference[T]) *low.ValueReference[T] {
	for n, o := range collection {
		if n.Value == item {
			return &o
		}
	}
	return nil
}

func ExtractSchema(root *yaml.Node) (*low.NodeReference[*Schema], error) {
	_, schLabel, schNode := utils.FindKeyNodeFull(SchemaLabel, root.Content)
	if schNode != nil {
		var schema Schema
		err := BuildModel(schNode, &schema)
		if err != nil {
			return nil, err
		}
		err = schema.Build(schNode, 0)
		if err != nil {
			return nil, err
		}
		return &low.NodeReference[*Schema]{Value: &schema, KeyNode: schLabel, ValueNode: schNode}, nil
	}
	return nil, nil
}

var mapLock sync.Mutex

func ExtractObjectRaw[T low.Buildable[N], N any](root *yaml.Node) (T, error) {
	var n T = new(N)
	err := BuildModel(root, n)
	if err != nil {
		return n, err
	}
	err = n.Build(root)
	if err != nil {
		return n, err
	}
	return n, nil
}

func ExtractObject[T low.Buildable[N], N any](label string, root *yaml.Node) (low.NodeReference[T], error) {
	_, ln, vn := utils.FindKeyNodeFull(label, root.Content)
	var n T = new(N)
	err := BuildModel(root, n)
	if err != nil {
		return low.NodeReference[T]{}, err
	}
	err = n.Build(root)
	if err != nil {
		return low.NodeReference[T]{}, err
	}
	return low.NodeReference[T]{
		Value:     n,
		KeyNode:   ln,
		ValueNode: vn,
	}, nil
}

func ExtractArray[T low.Buildable[N], N any](label string, root *yaml.Node) ([]low.NodeReference[T], *yaml.Node, *yaml.Node, error) {
	_, ln, vn := utils.FindKeyNodeFull(label, root.Content)
	var items []low.NodeReference[T]
	if vn != nil && ln != nil {
		for _, node := range vn.Content {
			var n T = new(N)
			err := BuildModel(node, n)
			if err != nil {
				return []low.NodeReference[T]{}, ln, vn, err
			}
			berr := n.Build(node)
			if berr != nil {
				return nil, ln, vn, berr
			}
			items = append(items, low.NodeReference[T]{
				Value:     n,
				ValueNode: node,
				KeyNode:   ln,
			})
		}
	}
	return items, ln, vn, nil
}

func ExtractMapFlat[PT low.Buildable[N], N any](label string, root *yaml.Node) (map[low.KeyReference[string]]low.ValueReference[PT], *yaml.Node, *yaml.Node, error) {
	_, labelNode, valueNode := utils.FindKeyNodeFull(label, root.Content)
	if valueNode != nil {
		var currentLabelNode *yaml.Node
		valueMap := make(map[low.KeyReference[string]]low.ValueReference[PT])
		for i, en := range valueNode.Content {
			if i%2 == 0 {
				currentLabelNode = en
				continue
			}
			if strings.HasPrefix(strings.ToLower(currentLabelNode.Value), "x-") {
				continue // yo, don't pay any attention to extensions, not here anyway.
			}
			var n PT = new(N)
			err := BuildModel(en, n)
			if err != nil {
				return nil, labelNode, valueNode, err
			}
			berr := n.Build(en)
			if berr != nil {
				return nil, labelNode, valueNode, berr
			}
			valueMap[low.KeyReference[string]{
				Value:   currentLabelNode.Value,
				KeyNode: currentLabelNode,
			}] = low.ValueReference[PT]{
				Value:     n,
				ValueNode: en,
			}
		}
		return valueMap, labelNode, valueNode, nil
	}
	return nil, labelNode, valueNode, nil
}

func ExtractMap[PT low.Buildable[N], N any](label string, root *yaml.Node) (map[low.KeyReference[string]]map[low.KeyReference[string]]low.ValueReference[PT], error) {
	_, labelNode, valueNode := utils.FindKeyNodeFull(label, root.Content)
	if valueNode != nil {
		var currentLabelNode *yaml.Node
		valueMap := make(map[low.KeyReference[string]]low.ValueReference[PT])
		for i, en := range valueNode.Content {
			if i%2 == 0 {
				currentLabelNode = en
				continue
			}
			if strings.HasPrefix(strings.ToLower(currentLabelNode.Value), "x-") {
				continue // yo, don't pay any attention to extensions, not here anyway.
			}
			var n PT = new(N)
			err := BuildModel(en, n)
			if err != nil {
				return nil, err
			}
			berr := n.Build(en)
			if berr != nil {
				return nil, berr
			}
			valueMap[low.KeyReference[string]{
				Value:   currentLabelNode.Value,
				KeyNode: currentLabelNode,
			}] = low.ValueReference[PT]{
				Value:     n,
				ValueNode: en,
			}
		}
		resMap := make(map[low.KeyReference[string]]map[low.KeyReference[string]]low.ValueReference[PT])
		resMap[low.KeyReference[string]{
			Value:   labelNode.Value,
			KeyNode: labelNode,
		}] = valueMap
		return resMap, nil
	}
	return nil, nil
}

func ExtractExtensions(root *yaml.Node) (map[low.KeyReference[string]]low.ValueReference[any], error) {
	extensions := utils.FindExtensionNodes(root.Content)
	extensionMap := make(map[low.KeyReference[string]]low.ValueReference[any])
	for _, ext := range extensions {
		if utils.IsNodeMap(ext.Value) {
			var v interface{}
			err := ext.Value.Decode(&v)
			if err != nil {
				return nil, err
			}
			extensionMap[low.KeyReference[string]{
				Value:   ext.Key.Value,
				KeyNode: ext.Key,
			}] = low.ValueReference[any]{Value: v, ValueNode: ext.Value}
		}
		if utils.IsNodeStringValue(ext.Value) {
			extensionMap[low.KeyReference[string]{
				Value:   ext.Key.Value,
				KeyNode: ext.Key,
			}] = low.ValueReference[any]{Value: ext.Value.Value, ValueNode: ext.Value}
		}
		if utils.IsNodeFloatValue(ext.Value) {
			fv, _ := strconv.ParseFloat(ext.Value.Value, 64)
			extensionMap[low.KeyReference[string]{
				Value:   ext.Key.Value,
				KeyNode: ext.Key,
			}] = low.ValueReference[any]{Value: fv, ValueNode: ext.Value}
		}
		if utils.IsNodeIntValue(ext.Value) {
			iv, _ := strconv.ParseInt(ext.Value.Value, 10, 64)
			extensionMap[low.KeyReference[string]{
				Value:   ext.Key.Value,
				KeyNode: ext.Key,
			}] = low.ValueReference[any]{Value: iv, ValueNode: ext.Value}
		}
		if utils.IsNodeBoolValue(ext.Value) {
			bv, _ := strconv.ParseBool(ext.Value.Value)
			extensionMap[low.KeyReference[string]{
				Value:   ext.Key.Value,
				KeyNode: ext.Key,
			}] = low.ValueReference[any]{Value: bv, ValueNode: ext.Value}
		}
		if utils.IsNodeArray(ext.Value) {
			var v []interface{}
			err := ext.Value.Decode(&v)
			if err != nil {
				return nil, err
			}
			extensionMap[low.KeyReference[string]{
				Value:   ext.Key.Value,
				KeyNode: ext.Key,
			}] = low.ValueReference[any]{Value: v, ValueNode: ext.Value}
		}
	}
	return extensionMap, nil
}