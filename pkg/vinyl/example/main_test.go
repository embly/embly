package main

// func TestBasic(t *testing.T) {
// 	t.Skip()
// 	db, err := vinyl.Connect("vinyl://max:password@localhost:8090/foo", vinyl.Metadata{
// 		Descriptor: proto.FileDescriptor("tables.proto"),
// 		Records: []vinyl.Record{{
// 			Name:       "User",
// 			PrimaryKey: "id",
// 			Indexes: []vinyl.Index{{
// 				Field:  "email",
// 				Unique: true,
// 			}},
// 		}},
// 	})
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	defer db.Close()
// 	user := User{
// 		Id:    "whatever",
// 		Email: "max@max.com",
// 	}
// 	fmt.Println("inserting")
// 	if err := db.Insert(&user); err != nil {
// 		t.Error(err)
// 	}
// 	fmt.Println("over")

// 	user2 := User{
// 		Id:    "whoever",
// 		Email: "max2@max.com",
// 	}
// 	if err := db.Insert(&user2); err != nil {
// 		t.Error(err)
// 	}

// 	queryResponse := []User{}
// 	if err := db.ExecuteQuery(&queryResponse,
// 		qm.Or(
// 			qm.Field("email").Equals("max@max.com"),
// 			qm.Field("email").Equals("foo@bar.com"),
// 		),
// 		qm.Limit(10),
// 	); err != nil {
// 		t.Error(err)
// 	}
// 	user.XXX_sizecache = 0
// 	user2.XXX_sizecache = 0
// 	assert.Equal(t, queryResponse, []User{user})

// 	users := []User{}
// 	if err := db.ExecuteQuery(&users, nil); err != nil {
// 		t.Error(err)
// 	}

// 	user.XXX_sizecache = 0
// 	user2.XXX_sizecache = 0
// 	assert.Equal(t, users, []User{user, user2})

// 	pkUser := User{}
// 	if err := db.LoadRecord(&pkUser, "whoever"); err != nil {
// 		t.Error(err)
// 	}
// 	assert.Equal(t, pkUser, user2)

// 	userAgain := User{}
// 	if err := db.DeleteRecord(&userAgain, "whoever"); err != nil {
// 		t.Error(err)
// 	}

// 	pkUser = User{}
// 	err = db.LoadRecord(&pkUser, "whoever")
// 	assert.Equal(t, pkUser, User{})

// 	if err := db.DeleteWhere(&User{}, nil); err != nil {
// 		t.Error(err)
// 	}

// 	users = []User{}
// 	_ = users
// 	if err := db.ExecuteQuery(&users, nil); err != nil {
// 		t.Error(err)
// 	}
// 	assert.Equal(t, []User{}, users)
// }
