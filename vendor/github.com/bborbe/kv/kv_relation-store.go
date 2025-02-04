// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

import (
	"context"

	"github.com/bborbe/errors"
)

//counterfeiter:generate -o mocks/relation-store.go --fake-name RelationStore . RelationStoreString
type RelationStoreString RelationStore[string, string]

// RelationStore implement a forward and backword id lookup for a 1:N relation.
type RelationStore[ID ~[]byte | ~string, RelatedID ~[]byte | ~string] interface {
	// Add the given relationIDs to ID
	Add(ctx context.Context, id ID, relatedIds []RelatedID) error
	// Replace all relations of id with the given
	Replace(ctx context.Context, id ID, relatedIds []RelatedID) error
	// Remove all relation from ID to the given
	Remove(ctx context.Context, id ID, relatedIds []RelatedID) error
	// Delete ID and all relations
	Delete(ctx context.Context, id ID) error
	// RelatedIDs return all relation of ID
	RelatedIDs(ctx context.Context, id ID) ([]RelatedID, error)
	// IDs return all ids of RelatedID
	IDs(ctx context.Context, relatedId RelatedID) ([]ID, error)
	// StreamIDs return all existings IDs
	StreamIDs(ctx context.Context, ch chan<- ID) error
	// StreamRelatedIDs return all existings relationIDs
	StreamRelatedIDs(ctx context.Context, ch chan<- RelatedID) error
	// MapIDRelations maps all entry to the given func
	MapIDRelations(ctx context.Context, fn func(ctx context.Context, key ID, relatedIDs []RelatedID) error) error
	// MapRelationIDs maps all entry to the given func
	MapRelationIDs(ctx context.Context, fn func(ctx context.Context, key RelatedID, ids []ID) error) error
}

func NewRelationStore[ID ~[]byte | ~string, RelatedID ~[]byte | ~string](db DB, name string) RelationStore[ID, RelatedID] {
	return &relationStore[ID, RelatedID]{
		relationStoreTx: NewRelationStoreTx[ID, RelatedID](name),
		db:              db,
	}
}

type relationStore[ID ~[]byte | ~string, RelatedID ~[]byte | ~string] struct {
	relationStoreTx RelationStoreTx[ID, RelatedID]
	db              DB
}

func (r *relationStore[ID, RelatedID]) MapIDRelations(ctx context.Context, fn func(ctx context.Context, key ID, relatedIDs []RelatedID) error) error {
	err := r.db.View(ctx, func(ctx context.Context, tx Tx) error {
		return r.relationStoreTx.MapIDRelations(ctx, tx, fn)
	})
	if err != nil {
		return errors.Wrapf(ctx, err, "view failed")
	}
	return nil
}

func (r *relationStore[ID, RelatedID]) MapRelationIDs(ctx context.Context, fn func(ctx context.Context, key RelatedID, ids []ID) error) error {
	err := r.db.View(ctx, func(ctx context.Context, tx Tx) error {
		return r.relationStoreTx.MapRelationIDs(ctx, tx, fn)
	})
	if err != nil {
		return errors.Wrapf(ctx, err, "view failed")
	}
	return nil
}

func (r *relationStore[ID, RelatedID]) StreamIDs(ctx context.Context, ch chan<- ID) error {
	err := r.db.View(ctx, func(ctx context.Context, tx Tx) error {
		return r.relationStoreTx.StreamIDs(ctx, tx, ch)
	})
	if err != nil {
		return errors.Wrapf(ctx, err, "view failed")
	}
	return nil
}

func (r *relationStore[ID, RelatedID]) StreamRelatedIDs(ctx context.Context, ch chan<- RelatedID) error {
	err := r.db.View(ctx, func(ctx context.Context, tx Tx) error {
		return r.relationStoreTx.StreamRelatedIDs(ctx, tx, ch)
	})
	if err != nil {
		return errors.Wrapf(ctx, err, "view failed")
	}
	return nil
}

func (r *relationStore[ID, RelatedID]) Add(ctx context.Context, id ID, relatedIds []RelatedID) error {
	err := r.db.Update(ctx, func(ctx context.Context, tx Tx) error {
		return r.relationStoreTx.Add(ctx, tx, id, relatedIds)
	})
	if err != nil {
		return errors.Wrapf(ctx, err, "update failed")
	}
	return nil
}

func (r *relationStore[ID, RelatedID]) Replace(ctx context.Context, id ID, relatedIds []RelatedID) error {
	err := r.db.Update(ctx, func(ctx context.Context, tx Tx) error {
		return r.relationStoreTx.Replace(ctx, tx, id, relatedIds)
	})
	if err != nil {
		return errors.Wrapf(ctx, err, "update failed")
	}
	return nil
}

func (r *relationStore[ID, RelatedID]) Remove(ctx context.Context, id ID, relatedIds []RelatedID) error {
	err := r.db.Update(ctx, func(ctx context.Context, tx Tx) error {
		return r.relationStoreTx.Remove(ctx, tx, id, relatedIds)
	})
	if err != nil {
		return errors.Wrapf(ctx, err, "update failed")
	}
	return nil
}

func (r *relationStore[ID, RelatedID]) Delete(ctx context.Context, id ID) error {
	err := r.db.Update(ctx, func(ctx context.Context, tx Tx) error {
		return r.relationStoreTx.Delete(ctx, tx, id)
	})
	if err != nil {
		return errors.Wrapf(ctx, err, "update failed")
	}
	return nil
}

func (r *relationStore[ID, RelatedID]) RelatedIDs(ctx context.Context, id ID) ([]RelatedID, error) {
	var result []RelatedID
	var err error
	err = r.db.View(ctx, func(ctx context.Context, tx Tx) error {
		result, err = r.relationStoreTx.RelatedIDs(ctx, tx, id)
		if err != nil {
			return errors.Wrapf(ctx, err, "TODO")
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "TODO")
	}
	return result, nil
}

func (r *relationStore[ID, RelatedID]) IDs(ctx context.Context, relatedId RelatedID) ([]ID, error) {
	var result []ID
	var err error
	err = r.db.View(ctx, func(ctx context.Context, tx Tx) error {
		result, err = r.relationStoreTx.IDs(ctx, tx, relatedId)
		if err != nil {
			return errors.Wrapf(ctx, err, "TODO")
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "TODO")
	}
	return result, nil
}
