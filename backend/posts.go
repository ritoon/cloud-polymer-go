// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package posts

import (
	"log"

	"appengine"
	"appengine/datastore"

	"github.com/GoogleCloudPlatform/go-endpoints/endpoints"
)

type PostsAPI struct{}

type Post struct {
	UID      *datastore.Key `json:"uid" datastore:"-"`
	Text     string         `json:"text"`
	Username string         `json:"username"`
	Avatar   string         `json:"avatar"`
	Favorite bool           `json:"favorite"`
}

type Posts struct {
	Posts []Post `json:"posts"`
}

func (PostsAPI) List(c endpoints.Context) (*Posts, error) {
	posts := []Post{}
	keys, err := datastore.NewQuery("Post").GetAll(c, &posts)
	if err != nil {
		return nil, err
	}
	for i, k := range keys {
		posts[i].UID = k
	}
	return &Posts{posts}, nil
}

type AddRequest struct {
	Text     string
	Username string
	Avatar   string
}

func (PostsAPI) Add(c endpoints.Context, r *AddRequest) (*Post, error) {
	k := datastore.NewIncompleteKey(c, "Post", nil)
	t := &Post{Text: r.Text, Username: r.Username, Avatar: r.Avatar}
	k, err := datastore.Put(c, k, t)
	if err != nil {
		return nil, err
	}
	t.UID = k
	return t, nil
}

type SetFavoriteRequest struct {
	UID      *datastore.Key
	Favorite bool
}

func (PostsAPI) SetFavorite(c endpoints.Context, r *SetFavoriteRequest) error {
	return datastore.RunInTransaction(c, func(c appengine.Context) error {
		var p Post
		if err := datastore.Get(c, r.UID, &p); err == datastore.ErrNoSuchEntity {
			return endpoints.NewNotFoundError("post not found")
		} else if err != nil {
			return err
		}
		p.Favorite = r.Favorite
		_, err := datastore.Put(c, r.UID, &p)
		return err
	}, nil)
}

func init() {
	api, err := endpoints.RegisterService(PostsAPI{}, "posts", "v1", "posts api", true)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(api.MethodByName("List"))
	info := api.MethodByName("List").Info()
	info.Name, info.HTTPMethod, info.Path = "getPosts", "GET", "posts"

	info = api.MethodByName("SetFavorite").Info()
	info.Name, info.HTTPMethod, info.Path = "setFavorite", "PUT", "posts"

	info = api.MethodByName("Add").Info()
	info.Name, info.HTTPMethod, info.Path = "addPost", "POST", "posts"

	endpoints.HandleHTTP()
}
