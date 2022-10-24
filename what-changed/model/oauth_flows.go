// Copyright 2022 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package model

import (
	"github.com/pb33f/libopenapi/datamodel/low"
	v3 "github.com/pb33f/libopenapi/datamodel/low/v3"
)

type OAuthFlowsChanges struct {
	PropertyChanges
	ImplicitChanges          *OAuthFlowChanges
	PasswordChanges          *OAuthFlowChanges
	ClientCredentialsChanges *OAuthFlowChanges
	AuthorizationCodeChanges *OAuthFlowChanges
	ExtensionChanges         *ExtensionChanges
}

func (o *OAuthFlowsChanges) TotalChanges() int {
	c := o.PropertyChanges.TotalChanges()
	if o.ImplicitChanges != nil {
		c += o.ImplicitChanges.TotalChanges()
	}
	if o.PasswordChanges != nil {
		c += o.PasswordChanges.TotalChanges()
	}
	if o.ClientCredentialsChanges != nil {
		c += o.ClientCredentialsChanges.TotalChanges()
	}
	if o.AuthorizationCodeChanges != nil {
		c += o.AuthorizationCodeChanges.TotalChanges()
	}
	if o.ExtensionChanges != nil {
		c += o.ExtensionChanges.TotalChanges()
	}
	return c
}

func (o *OAuthFlowsChanges) TotalBreakingChanges() int {
	c := o.PropertyChanges.TotalBreakingChanges()
	if o.ImplicitChanges != nil {
		c += o.ImplicitChanges.TotalBreakingChanges()
	}
	if o.PasswordChanges != nil {
		c += o.PasswordChanges.TotalBreakingChanges()
	}
	if o.ClientCredentialsChanges != nil {
		c += o.ClientCredentialsChanges.TotalBreakingChanges()
	}
	if o.AuthorizationCodeChanges != nil {
		c += o.AuthorizationCodeChanges.TotalBreakingChanges()
	}
	return c
}

func CompareOAuthFlows(l, r *v3.OAuthFlows) *OAuthFlowsChanges {
	if low.AreEqual(l, r) {
		return nil
	}

	oa := new(OAuthFlowsChanges)
	var changes []*Change

	// client credentials
	if !l.ClientCredentials.IsEmpty() && !r.ClientCredentials.IsEmpty() {
		oa.ClientCredentialsChanges = CompareOAuthFlow(l.ClientCredentials.Value, r.ClientCredentials.Value)
	}
	if !l.ClientCredentials.IsEmpty() && r.ClientCredentials.IsEmpty() {
		CreateChange(&changes, ObjectRemoved, v3.ClientCredentialsLabel,
			l.ClientCredentials.ValueNode, nil, true,
			l.ClientCredentials.Value, nil)
	}
	if l.ClientCredentials.IsEmpty() && !r.ClientCredentials.IsEmpty() {
		CreateChange(&changes, ObjectAdded, v3.ClientCredentialsLabel,
			nil, r.ClientCredentials.ValueNode, false,
			nil, r.ClientCredentials.Value)
	}

	// implicit
	if !l.Implicit.IsEmpty() && !r.Implicit.IsEmpty() {
		oa.ImplicitChanges = CompareOAuthFlow(l.Implicit.Value, r.Implicit.Value)
	}
	if !l.Implicit.IsEmpty() && r.Implicit.IsEmpty() {
		CreateChange(&changes, ObjectRemoved, v3.ImplicitLabel,
			l.Implicit.ValueNode, nil, true,
			l.Implicit.Value, nil)
	}
	if l.Implicit.IsEmpty() && !r.Implicit.IsEmpty() {
		CreateChange(&changes, ObjectAdded, v3.ImplicitLabel,
			nil, r.Implicit.ValueNode, false,
			nil, r.Implicit.Value)
	}

	// password
	if !l.Password.IsEmpty() && !r.Password.IsEmpty() {
		oa.PasswordChanges = CompareOAuthFlow(l.Password.Value, r.Password.Value)
	}
	if !l.Password.IsEmpty() && r.Password.IsEmpty() {
		CreateChange(&changes, ObjectRemoved, v3.PasswordLabel,
			l.Password.ValueNode, nil, true,
			l.Password.Value, nil)
	}
	if l.Password.IsEmpty() && !r.Password.IsEmpty() {
		CreateChange(&changes, ObjectAdded, v3.PasswordLabel,
			nil, r.Password.ValueNode, false,
			nil, r.Password.Value)
	}

	// auth code
	if !l.AuthorizationCode.IsEmpty() && !r.AuthorizationCode.IsEmpty() {
		oa.AuthorizationCodeChanges = CompareOAuthFlow(l.AuthorizationCode.Value, r.AuthorizationCode.Value)
	}
	if !l.AuthorizationCode.IsEmpty() && r.AuthorizationCode.IsEmpty() {
		CreateChange(&changes, ObjectRemoved, v3.AuthorizationCodeLabel,
			l.AuthorizationCode.ValueNode, nil, true,
			l.AuthorizationCode.Value, nil)
	}
	if l.AuthorizationCode.IsEmpty() && !r.AuthorizationCode.IsEmpty() {
		CreateChange(&changes, ObjectAdded, v3.AuthorizationCodeLabel,
			nil, r.AuthorizationCode.ValueNode, false,
			nil, r.AuthorizationCode.Value)
	}
	oa.ExtensionChanges = CompareExtensions(l.Extensions, r.Extensions)
	oa.Changes = changes
	return oa
}

type OAuthFlowChanges struct {
	PropertyChanges
	ExtensionChanges *ExtensionChanges
}

func (o *OAuthFlowChanges) TotalChanges() int {
	c := o.PropertyChanges.TotalChanges()
	if o.ExtensionChanges != nil {
		c += o.ExtensionChanges.TotalChanges()
	}
	return c
}

func (o *OAuthFlowChanges) TotalBreakingChanges() int {
	return o.PropertyChanges.TotalBreakingChanges()
}

func CompareOAuthFlow(l, r *v3.OAuthFlow) *OAuthFlowChanges {
	if low.AreEqual(l, r) {
		return nil
	}

	var changes []*Change
	var props []*PropertyCheck

	// authorization url
	props = append(props, &PropertyCheck{
		LeftNode:  l.AuthorizationUrl.ValueNode,
		RightNode: r.AuthorizationUrl.ValueNode,
		Label:     v3.AuthorizationUrlLabel,
		Changes:   &changes,
		Breaking:  true,
		Original:  l,
		New:       r,
	})

	// token url
	props = append(props, &PropertyCheck{
		LeftNode:  l.TokenUrl.ValueNode,
		RightNode: r.TokenUrl.ValueNode,
		Label:     v3.TokenUrlLabel,
		Changes:   &changes,
		Breaking:  true,
		Original:  l,
		New:       r,
	})

	// refresh url
	props = append(props, &PropertyCheck{
		LeftNode:  l.RefreshUrl.ValueNode,
		RightNode: r.RefreshUrl.ValueNode,
		Label:     v3.RefreshUrlLabel,
		Changes:   &changes,
		Breaking:  true,
		Original:  l,
		New:       r,
	})

	CheckProperties(props)

	for v := range l.Scopes.Value {
		if r != nil && r.FindScope(v.Value) == nil {
			CreateChange(&changes, ObjectRemoved, v3.Scopes,
				l.Scopes.Value[v].ValueNode, nil, true,
				v.Value, nil)
			continue
		}
		if r != nil && r.FindScope(v.Value) != nil {
			if l.Scopes.Value[v].Value != r.FindScope(v.Value).Value {
				CreateChange(&changes, Modified, v3.Scopes,
					l.Scopes.Value[v].ValueNode, r.FindScope(v.Value).ValueNode, true,
					l.Scopes.Value[v].Value, r.FindScope(v.Value).Value)
			}
		}
	}
	for v := range r.Scopes.Value {
		if l != nil && l.FindScope(v.Value) == nil {
			CreateChange(&changes, ObjectAdded, v3.Scopes,
				nil, r.Scopes.Value[v].ValueNode, false,
				nil, v.Value)
		}
	}
	oa := new(OAuthFlowChanges)
	oa.Changes = changes
	oa.ExtensionChanges = CompareExtensions(l.Extensions, r.Extensions)
	return oa
}
