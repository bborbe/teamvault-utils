// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

import (
	"context"

	"github.com/bborbe/errors"
)

//counterfeiter:generate -o mocks/relation-store-tx.go --fake-name RelationStoreTx . RelationStoreTxString
type RelationStoreTxString RelationStoreTx[string, string]

type RelationStoreTx[ID ~[]byte | ~string, RelatedID ~[]byte | ~string] interface {
	// Add the given relationIDs to ID
	Add(ctx context.Context, tx Tx, id ID, relatedIds []RelatedID) error
	// Replace all relations of id with the given
	Replace(ctx context.Context, tx Tx, id ID, relatedIds []RelatedID) error
	// Remove all relation from ID to the given
	Remove(ctx context.Context, tx Tx, id ID, relatedIds []RelatedID) error
	// Delete ID and all relations
	Delete(ctx context.Context, tx Tx, id ID) error
	// RelatedIDs return all relation of ID
	RelatedIDs(ctx context.Context, tx Tx, id ID) ([]RelatedID, error)
	// IDs return all ids of RelatedID
	IDs(ctx context.Context, tx Tx, relatedId RelatedID) ([]ID, error)
	// StreamIDs return all existings IDs
	StreamIDs(ctx context.Context, tx Tx, ch chan<- ID) error
	// StreamRelatedIDs return all existings relationIDs
	StreamRelatedIDs(ctx context.Context, tx Tx, ch chan<- RelatedID) error
	// MapIDRelations maps all entry to the given func
	MapIDRelations(ctx context.Context, tx Tx, fn func(ctx context.Context, key ID, relatedIDs []RelatedID) error) error
	// MapRelationIDs maps all entry to the given func
	MapRelationIDs(ctx context.Context, tx Tx, fn func(ctx context.Context, key RelatedID, ids []ID) error) error
}

func NewRelationStoreTx[ID ~[]byte | ~string, RelatedID ~[]byte | ~string](name string) RelationStoreTx[ID, RelatedID] {
	return &relationStoreTx[ID, RelatedID]{
		relationIdBucket: NewStoreTx[RelatedID, []ID](BucketFromStrings(name, "relation", "id")),
		idRelationBucket: NewStoreTx[ID, []RelatedID](BucketFromStrings(name, "id", "relation")),
	}
}

type relationStoreTx[ID ~[]byte | ~string, RelatedID ~[]byte | ~string] struct {
	idRelationBucket StoreTx[ID, []RelatedID]
	relationIdBucket StoreTx[RelatedID, []ID]
}

func (r relationStoreTx[ID, RelatedID]) MapIDRelations(ctx context.Context, tx Tx, fn func(ctx context.Context, key ID, relatedIDs []RelatedID) error) error {
	return r.idRelationBucket.Map(ctx, tx, fn)
}

func (r relationStoreTx[ID, RelatedID]) MapRelationIDs(ctx context.Context, tx Tx, fn func(ctx context.Context, key RelatedID, ids []ID) error) error {
	return r.relationIdBucket.Map(ctx, tx, fn)
}

func (r relationStoreTx[ID, RelatedID]) Add(ctx context.Context, tx Tx, id ID, relatedIds []RelatedID) error {
	currentRelationIDs, err := r.RelatedIDs(ctx, tx, id)
	if err != nil {
		return errors.Wrapf(ctx, err, "get relationIDs failed")
	}
	currentRelationIDs = unique(append(currentRelationIDs, relatedIds...))
	if err := r.idRelationBucket.Add(ctx, tx, id, currentRelationIDs); err != nil {
		return errors.Wrapf(ctx, err, "add relationIDs failed")
	}
	for _, relatedId := range relatedIds {
		currentIDs, err := r.IDs(ctx, tx, relatedId)
		if err != nil {
			return errors.Wrapf(ctx, err, "get ids failed")
		}
		currentIDs = unique(append(currentIDs, id))
		if err := r.relationIdBucket.Add(ctx, tx, relatedId, currentIDs); err != nil {
			return errors.Wrapf(ctx, err, "add ids failed")
		}
	}
	return nil
}

func (r relationStoreTx[ID, RelatedID]) Remove(ctx context.Context, tx Tx, id ID, relatedIds []RelatedID) error {
	currentRelationIDs, err := r.RelatedIDs(ctx, tx, id)
	if err != nil {
		return errors.Wrapf(ctx, err, "get relationIDs failed")
	}
	currentRelationIDs = remove(currentRelationIDs, relatedIds...)
	if err := r.idRelationBucket.Add(ctx, tx, id, currentRelationIDs); err != nil {
		return errors.Wrapf(ctx, err, "add relationIDs failed")
	}
	for _, relatedId := range relatedIds {
		currentIDs, err := r.IDs(ctx, tx, relatedId)
		if err != nil {
			return errors.Wrapf(ctx, err, "get ids failed")
		}
		currentIDs = remove(currentIDs, id)
		if err := r.relationIdBucket.Add(ctx, tx, relatedId, currentIDs); err != nil {
			return errors.Wrapf(ctx, err, "add ids failed")
		}
	}
	return nil
}

func (r relationStoreTx[ID, RelatedID]) Delete(ctx context.Context, tx Tx, id ID) error {
	relatedIDs, err := r.RelatedIDs(ctx, tx, id)
	if err != nil {
		return errors.Wrapf(ctx, err, "get relationIDs for id %s failed", id)
	}
	if err := r.Remove(ctx, tx, id, relatedIDs); err != nil {
		return errors.Wrapf(ctx, err, "remove relationIDs for id %s failed", id)
	}
	if err := r.idRelationBucket.Remove(ctx, tx, id); err != nil {
		return errors.Wrapf(ctx, err, "remove id %s failed", id)
	}
	return nil
}

func (r relationStoreTx[ID, RelatedID]) Replace(ctx context.Context, tx Tx, id ID, relatedIds []RelatedID) error {
	if err := r.Delete(ctx, tx, id); err != nil {
		return err
	}
	if err := r.Add(ctx, tx, id, relatedIds); err != nil {
		return err
	}
	return nil
}

func (r relationStoreTx[ID, RelatedID]) RelatedIDs(ctx context.Context, tx Tx, id ID) ([]RelatedID, error) {
	result, err := r.idRelationBucket.Get(ctx, tx, id)
	if err != nil {
		if errors.Is(err, BucketNotFoundError) || errors.Is(err, KeyNotFoundError) {
			return nil, nil
		}
		return nil, errors.Wrapf(ctx, err, "get failed")
	}
	return *result, nil
}

func (r relationStoreTx[ID, RelatedID]) IDs(ctx context.Context, tx Tx, relatedId RelatedID) ([]ID, error) {
	result, err := r.relationIdBucket.Get(ctx, tx, relatedId)
	if err != nil {
		if errors.Is(err, BucketNotFoundError) || errors.Is(err, KeyNotFoundError) {
			return nil, nil
		}
		return nil, errors.Wrapf(ctx, err, "get failed")
	}
	return *result, nil
}

func (r relationStoreTx[ID, RelatedID]) StreamIDs(ctx context.Context, tx Tx, ch chan<- ID) error {
	err := r.idRelationBucket.Map(ctx, tx, func(ctx context.Context, key ID, object []RelatedID) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- key:
			return nil
		}
	})
	if err != nil {
		if errors.Is(err, BucketNotFoundError) || errors.Is(err, KeyNotFoundError) {
			return nil
		}
		return errors.Wrapf(ctx, err, "map failed")
	}
	return nil
}

func (r relationStoreTx[ID, RelatedID]) StreamRelatedIDs(ctx context.Context, tx Tx, ch chan<- RelatedID) error {
	err := r.relationIdBucket.Map(ctx, tx, func(ctx context.Context, key RelatedID, object []ID) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- key:
			return nil
		}
	})
	if err != nil {
		if errors.Is(err, BucketNotFoundError) || errors.Is(err, KeyNotFoundError) {
			return nil
		}
		return errors.Wrapf(ctx, err, "map failed")
	}
	return nil
}

func unique[T ~[]byte | ~string](list []T) []T {
	result := make([]T, 0)
	found := make(map[string]bool)
	for _, l := range list {
		if found[string(l)] {
			continue
		}
		found[string(l)] = true
		result = append(result, l)
	}
	return result
}

func remove[T ~[]byte | ~string](list []T, removes ...T) []T {
	found := make(map[string]bool)
	for _, l := range removes {
		found[string(l)] = true
	}
	result := make([]T, 0)
	for _, l := range list {
		if found[string(l)] {
			continue
		}
		found[string(l)] = true
		result = append(result, l)
	}
	return result
}
