package feed

import (
	"servitor/object"
	"servitor/pub"
	"testing"
)

var post1, _ = pub.NewPostFromObject(object.Object{
	"type":    "Note",
	"content": "Here from post1",
}, nil)

var post2, _ = pub.NewPostFromObject(object.Object{
	"type":    "Video",
	"content": "Here from post2",
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
	shouldBePost1 := feed.Get(0)
	if shouldBePost1 != post1 {
		t.Fatalf("Posts differed after create centerless, is %#v but should be %#v", shouldBePost1, post1)
	}
	defer func() {
		if recover() == nil {
			t.Fatalf("After create centerless, Get(0) should have panicked but did not")
		}
	}()
	feed.Get(-1)
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

func TestMoveDown(t *testing.T) {
	feed := CreateAndAppend([]pub.Tangible{post1, post2})
	feed.MoveDown()
	shouldBePost2 := feed.Current()
	if shouldBePost2 != post2 {
		t.Fatalf("is %#v but should be %#v", shouldBePost2, post2)
	}
	shouldBePost1 := feed.Get(-1)
	if shouldBePost1 != post1 {
		t.Fatalf("is %#v but should be %#v", shouldBePost1, post1)
	}

	defer func() {
		if recover() == nil {
			t.Fatalf("Get(1) should have panicked but did not")
		}
	}()
	feed.Get(1)

	/* Repeat everything exactly */
	feed.MoveDown()
	shouldBePost2 = feed.Current()
	if shouldBePost2 != post2 {
		t.Fatalf("is %#v but should be %#v", shouldBePost2, post2)
	}
	shouldBePost1 = feed.Get(-1)
	if shouldBePost1 != post1 {
		t.Fatalf("is %#v but should be %#v", shouldBePost1, post1)
	}

	defer func() {
		if recover() == nil {
			t.Fatalf("Get(1) should have panicked but did not")
		}
	}()
	feed.Get(1)
}

func TestMoveUp(t *testing.T) {
	feed := Create(post1)
	feed.Prepend([]pub.Tangible{post2})
	feed.MoveUp()
	shouldBePost2 := feed.Current()
	if shouldBePost2 != post2 {
		t.Fatalf("is %#v but should be %#v", shouldBePost2, post2)
	}
	shouldBePost1 := feed.Get(1)
	if shouldBePost1 != post1 {
		t.Fatalf("is %#v but should be %#v", shouldBePost1, post1)
	}

	defer func() {
		if recover() == nil {
			t.Fatalf("Get(-1) should have panicked but did not")
		}
	}()
	feed.Get(-1)

	/* Repeat everything exactly */
	feed.MoveUp()
	shouldBePost2 = feed.Current()
	if shouldBePost2 != post2 {
		t.Fatalf("is %#v but should be %#v", shouldBePost2, post2)
	}
	shouldBePost1 = feed.Get(1)
	if shouldBePost1 != post1 {
		t.Fatalf("is %#v but should be %#v", shouldBePost1, post1)
	}

	defer func() {
		if recover() == nil {
			t.Fatalf("Get(-1) should have panicked but did not")
		}
	}()
	feed.Get(-1)
}
