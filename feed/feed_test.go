package feed

import (
	"testing"
	"mimicry/pub"
	"mimicry/object"
)

var post1, _ = pub.NewPostFromObject(object.Object {
	"type": "Note",
	"content": "Hello!",
}, nil)

var post2, _ = pub.NewPostFromObject(object.Object {
	"type": "Video",
	"content": "Goodbye!",
}, nil)

func TestCreate(t *testing.T) {
	feed := Create(post1)
	shouldBePost1 := feed.Get(0)
	if shouldBePost1 != post1 {
		t.Fatalf("Center Posts differ after Create, is %#v but should be %#v", shouldBePost1, post1)
	}
}

func TestCreateCreateAndAppend(t *testing.T) {
	feed := CreateAndAppend([]pub.Tangible{post1})
	shouldBePost1 := feed.Get(1)
	if shouldBePost1 != post1 {
		t.Fatalf("Posts differed after create centerless, is %#v but should be %#v", shouldBePost1, post1)
	}
	defer func() {
        if recover() == nil {
            t.Fatalf("After create centerless, Get(0) should have panicked but did not")
        }
    }()
	feed.Get(0)
}

func TestAppend(t *testing.T) {
	feed := Create(post1)
	feed.Append([]pub.Tangible{post2})
	shouldBePost1 := feed.Get(0)
	shouldBePost2 := feed.Get(1)
	if shouldBePost1 != post1 {
		t.Fatalf("Center Posts differ after Append, is %#v but should be %#v", shouldBePost1, post1)
	}
	if shouldBePost2 != post2 {
		t.Fatalf("Appended posts differ, is %#v but should be %#v", shouldBePost2, post2)
	}
}

func TestPrepend(t *testing.T) {
	feed := Create(post1)
	feed.Prepend([]pub.Tangible{post2})
	shouldBePost1 := feed.Get(0)
	shouldBePost2 := feed.Get(-1)
	if shouldBePost1 != post1 {
		t.Fatalf("Center Posts differ after Prepend, is %#v but should be %#v", shouldBePost1, post1)
	}
	if shouldBePost2 != post2 {
		t.Fatalf("Prepended posts differ, is %#v but should be %#v", shouldBePost2, post2)
	}
}