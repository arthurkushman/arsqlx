package buildsqlx

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestTable    = "test"
	PostsTable   = "posts"
	UsersTable   = "users"
	TestUserName = "Dead Beaf"
)

var (
	dbConnInfo = fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "postgres", "postgres", "postgres")

	dataMap = map[string]interface{}{"foo": "foo foo foo", "bar": "bar bar bar", "baz": int64(123)}

	db = NewDb(NewConnection("postgres", dbConnInfo))
)

func TestMain(m *testing.M) {
	_, err := db.Sql().Exec("create table if not exists users (id serial primary key, name varchar(128) not null, points integer)")
	if err != nil {
		panic(err)
	}

	_, err = db.Sql().Exec("create table if not exists posts (id serial primary key, title varchar(128) not null, post text, user_id integer, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW())")
	if err != nil {
		panic(err)
	}

	_, err = db.Sql().Exec("create table if not exists test (id serial primary key, foo varchar(128) not null, bar varchar(128) not null, baz integer)")
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestNewConnectionFromDB(t *testing.T) {
	conn := NewConnectionFromDb(&sql.DB{})
	require.Equal(t, conn.db, &sql.DB{})
}

func TestSelectAndLimit(t *testing.T) {
	_, err := db.Truncate(TestTable)
	assert.NoError(t, err)

	err = db.Table(TestTable).Insert(dataMap)
	assert.NoError(t, err)

	qDb := db.Table(TestTable).Select("foo", "bar")
	res, err := qDb.AddSelect("baz").Limit(15).Get()
	assert.NoError(t, err)

	for k, mapVal := range dataMap {
		for _, v := range res {
			assert.Equal(t, v[k], mapVal)
		}
	}

	_, err = db.Truncate(TestTable)
	assert.NoError(t, err)
}

func TestInsert(t *testing.T) {
	_, err := db.Truncate(TestTable)
	assert.NoError(t, err)

	err = db.Table(TestTable).Insert(dataMap)
	assert.NoError(t, err)

	res, err := db.Table(TestTable).Select("foo", "bar", "baz").Get()
	assert.NoError(t, err)

	for k, mapVal := range dataMap {
		for _, v := range res {
			assert.Equal(t, v[k], mapVal)
		}
	}

	_, err = db.Truncate(TestTable)
	assert.NoError(t, err)
}

var batchData = []map[string]interface{}{
	0: {"foo": "foo foo foo", "bar": "bar bar bar", "baz": int64(123)},
	1: {"foo": "foo foo foo foo", "bar": "bar bar bar bar", "baz": int64(1234)},
	2: {"foo": "foo foo foo foo foo", "bar": "bar bar bar bar bar", "baz": int64(12345)},
}

func TestInsertBatchSelectMultiple(t *testing.T) {
	_, err := db.Truncate(TestTable)
	assert.NoError(t, err)

	err = db.Table(TestTable).InsertBatch(batchData)
	assert.NoError(t, err)

	res, err := db.Table(TestTable).Select("foo", "bar", "baz").OrderBy("foo", "ASC").Get()
	assert.NoError(t, err)

	for mapKey, mapVal := range batchData {
		for k, mV := range mapVal {
			assert.Equal(t, res[mapKey][k], mV)
		}
	}

	_, err = db.Truncate(TestTable)
	assert.NoError(t, err)
}

func TestWhereOnly(t *testing.T) {
	var cmp = "foo foo foo"

	_, err := db.Truncate(TestTable)
	assert.NoError(t, err)

	err = db.Table(TestTable).InsertBatch(batchData)
	assert.NoError(t, err)
	res, err := db.Table(TestTable).Select("foo", "bar", "baz").Where("foo", "=", cmp).Get()
	assert.NoError(t, err)

	assert.Equal(t, res[0]["foo"], cmp)

	_, err = db.Truncate(TestTable)
	assert.NoError(t, err)
}

func TestWhereAndOr(t *testing.T) {
	var cmp = "foo foo foo"

	_, err := db.Truncate(TestTable)
	assert.NoError(t, err)

	err = db.Table(TestTable).InsertBatch(batchData)
	assert.NoError(t, err)
	res, err := db.Table(TestTable).Select("foo", "bar", "baz").Where("foo", "=", cmp).AndWhere("bar", "!=", "foo").OrWhere("baz", "=", 123).Get()
	assert.NoError(t, err)

	assert.Equal(t, res[0]["foo"], cmp)

	_, err = db.Truncate(TestTable)
	assert.NoError(t, err)
}

//var users = `create table users (id serial primary key, name varchar(128) not null, points integer)`
//
//var posts = `create table posts (id serial primary key, title varchar(128) not null, post text, user_id integer, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`

var batchUsers = []map[string]interface{}{
	0: {"id": int64(1), "name": "Alex Shmidt", "points": int64(123)},
	1: {"id": int64(2), "name": "Darth Vader", "points": int64(1234)},
	2: {"id": int64(3), "name": "Dead Beaf", "points": int64(12345)},
	3: {"id": int64(4), "name": "Dead Beaf", "points": int64(12345)},
}

var batchPosts = []map[string]interface{}{
	0: {"id": int64(1), "title": "Lorem ipsum dolor sit amet,", "post": "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", "user_id": int64(1), "updated_at": "2086-09-09 18:27:40"},
	1: {"id": int64(2), "title": "Sed ut perspiciatis unde omnis iste natus", "post": "Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo.", "user_id": int64(2), "updated_at": "2087-09-09 18:27:40"},
	2: {"id": int64(3), "title": "Ut enim ad minima veniam", "post": "Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit laboriosam, nisi ut aliquid ex ea commodi consequatur?", "user_id": int64(3), "updated_at": "2088-09-09 18:27:40"},
	3: {"id": int64(4), "title": "Lorem ipsum dolor sit amet,", "post": nil, "user_id": nil, "updated_at": "2086-09-09 18:27:40"},
}

func TestJoins(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	_, err = db.Truncate(PostsTable)
	assert.NoError(t, err)

	var posts []map[string]interface{}
	for _, v := range batchUsers {
		id, err := db.Table(UsersTable).InsertGetId(v)
		assert.NoError(t, err)

		posts = append(posts, map[string]interface{}{
			"title": "ttl", "post": "foo bar baz", "user_id": id,
		})
	}

	err = db.Table(PostsTable).InsertBatch(posts)
	assert.NoError(t, err)

	res, err := db.Table(UsersTable).Select("name", "post", "user_id").LeftJoin("posts", "users.id", "=", "posts.user_id").Get()
	assert.NoError(t, err)

	for k, val := range res {
		assert.Equal(t, val["name"], batchUsers[k]["name"])
		assert.Equal(t, val["user_id"], batchUsers[k]["id"])
	}

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)

	_, err = db.Truncate(PostsTable)
	assert.NoError(t, err)
}

var rowsToUpdate = []struct {
	insert map[string]interface{}
	update map[string]interface{}
}{
	{map[string]interface{}{"foo": "foo foo foo", "bar": "bar bar bar", "baz": 123}, map[string]interface{}{"foo": "foo changed", "baz": nil}},
}

func TestUpdate(t *testing.T) {
	_, err := db.Truncate(TestTable)
	assert.NoError(t, err)

	for _, obj := range rowsToUpdate {
		err := db.Table(TestTable).Insert(obj.insert)
		assert.NoError(t, err)

		rows, err := db.Table(TestTable).Where("foo", "=", "foo foo foo").Update(obj.update)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, rows, int64(1))

		res, err := db.Table(TestTable).Select("foo").Where("foo", "=", obj.update["foo"]).Get()
		assert.NoError(t, err)
		assert.Equal(t, obj.update["foo"], res[0]["foo"])
	}

	_, err = db.Truncate(TestTable)
	assert.NoError(t, err)
}

var rowsToDelete = []struct {
	insert map[string]interface{}
	where  map[string]interface{}
}{
	{map[string]interface{}{"foo": "foo foo foo", "bar": "bar bar bar", "baz": 123}, map[string]interface{}{"bar": 123}},
}

func TestDelete(t *testing.T) {
	_, err := db.Truncate(TestTable)
	assert.NoError(t, err)

	for _, obj := range rowsToDelete {
		err := db.Table(TestTable).Insert(obj.insert)
		assert.NoError(t, err)

		rows, err := db.Table(TestTable).Where("baz", "=", obj.where["bar"]).Delete()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, rows, int64(1))
	}
}

var incrDecr = []struct {
	insert  map[string]interface{}
	incr    uint64
	incrRes uint64
	decr    uint64
	decrRes uint64
}{
	{map[string]interface{}{"foo": "foo foo foo", "bar": "bar bar bar", "baz": 1}, 3, 4, 1, 3},
}

func TestDB_Increment_Decrement(t *testing.T) {
	_, err := db.Truncate(TestTable)
	assert.NoError(t, err)

	for _, obj := range incrDecr {
		err = db.Table(TestTable).Insert(obj.insert)
		assert.NoError(t, err)

		_, err = db.Table(TestTable).Increment("baz", obj.incr)
		assert.NoError(t, err)

		res, err := db.Table(TestTable).Select("baz").Where("baz", "=", obj.incrRes).Get()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(res), 1)
		assert.Equal(t, res[0]["baz"], int64(obj.incrRes))

		_, err = db.Table(TestTable).Decrement("baz", obj.decr)
		assert.NoError(t, err)

		res, err = db.Table(TestTable).Select("baz").Where("baz", "=", obj.decrRes).Get()
		assert.NoError(t, err)

		assert.GreaterOrEqual(t, len(res), 1)
		assert.Equal(t, res[0]["baz"], int64(obj.decrRes))
	}

	_, err = db.Truncate(TestTable)
	assert.NoError(t, err)
}

var rowsToReplace = []struct {
	insert   map[string]interface{}
	conflict string
	replace  map[string]interface{}
}{
	{map[string]interface{}{"id": 1, "foo": "foo foo foo", "bar": "bar bar bar", "baz": 123}, "id", map[string]interface{}{"id": 1, "foo": "baz baz baz", "bar": "bar bar bar", "baz": 123}},
}

func TestDB_Replace(t *testing.T) {
	_, err := db.Truncate(TestTable)
	assert.NoError(t, err)

	for _, obj := range rowsToReplace {
		_, err := db.Table(TestTable).Replace(obj.insert, obj.conflict)
		assert.NoError(t, err)

		rows, err := db.Table(TestTable).Replace(obj.replace, obj.conflict)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, rows, int64(1))

		res, err := db.Table(TestTable).Select("foo").Where("baz", "=", obj.replace["baz"]).Get()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(res), 1)
		assert.Equal(t, res[0]["foo"], obj.replace["foo"])
	}

	_, err = db.Truncate(TestTable)
	assert.NoError(t, err)
}

var userForUnion = map[string]interface{}{"id": int64(1), "name": "Alex Shmidt", "points": int64(123)}

func TestDB_Union(t *testing.T) {
	_, err := db.Truncate(TestTable)
	assert.NoError(t, err)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(TestTable).Insert(dataMap)
	assert.NoError(t, err)
	err = db.Table(UsersTable).Insert(userForUnion)
	assert.NoError(t, err)

	union := db.Table(TestTable).Select("bar", "baz").Union()
	res, err := union.Table(UsersTable).Select("name", "points").Get()
	assert.NoError(t, err)
	for _, v := range res {
		assert.Equal(t, v["baz"], userForUnion["points"])
	}

	_, err = db.Truncate(TestTable)
	assert.NoError(t, err)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_InTransaction(t *testing.T) {
	var tests = map[string]struct {
		dataMap map[string]interface{}
		res     interface{}
		err     error
	}{
		"transaction commit ok": {
			dataMap: dataMap,
			res:     1,
			err:     nil,
		},
		"transaction commit ok int64": {
			dataMap: dataMap,
			res:     int64(1),
			err:     nil,
		},
		"transaction commit ok uint64": {
			dataMap: dataMap,
			res:     uint64(1),
			err:     nil,
		},
		"transaction commit ok map[string]interface{}": {
			dataMap: dataMap,
			res:     map[string]interface{}{"foo": "foo foo foo", "bar": "bar bar bar", "baz": int64(123)},
			err:     nil,
		},
		"transaction commit ok []map[string]interface{}": {
			dataMap: dataMap,
			res: []map[string]interface{}{
				{
					"foo": "foo foo foo", "bar": "bar bar bar", "baz": int64(123),
				},
			},
			err: nil,
		},
		"transaction early exit err": {
			dataMap: dataMap,
			res:     0,
			err:     errors.New("some error"),
		},
		"transaction rollback": {
			dataMap: dataMap,
			res:     0,
			err:     nil,
		},
	}

	for n, tt := range tests {
		t.Run(n, func(t *testing.T) {
			_, err := db.Truncate(TestTable)
			assert.NoError(t, err)

			defer func() {
				_, err = db.Truncate(TestTable)
				assert.NoError(t, err)
			}()

			err = db.InTransaction(func() (any, error) {
				err = db.Table(TestTable).Insert(tt.dataMap)

				return tt.res, tt.err
			})

			if tt.err != nil {
				assert.Error(t, tt.err, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDB_HasTable(t *testing.T) {
	tblExists, err := db.HasTable("public", PostsTable)
	assert.NoError(t, err)
	assert.True(t, tblExists)
}

func TestDB_HasColumns(t *testing.T) {
	colsExists, err := db.HasColumns("public", PostsTable, "title", "user_id")
	assert.NoError(t, err)
	assert.True(t, colsExists)
}

func TestDB_First(t *testing.T) {
	_, err := db.Truncate(TestTable)
	assert.NoError(t, err)

	err = db.Table(TestTable).Insert(dataMap)
	assert.NoError(t, err)

	// write concurrent row to order and get the only 1st
	err = db.Table(TestTable).Insert(map[string]interface{}{"foo": "foo foo foo 2", "bar": "bar bar bar 2", "baz": int64(1234)})
	assert.NoError(t, err)

	res, err := db.Table(TestTable).Select("baz").OrderBy("baz", "desc").OrderBy("foo", "desc").First()
	assert.NoError(t, err)
	assert.Equal(t, res["baz"], int64(1234))

	_, err = db.Table(TestTable).Select("baz").OrderBy("baz", "desc").OrderBy("fo", "desc").First()
	assert.Error(t, err)

	_, err = db.Truncate(TestTable)
	assert.NoError(t, err)
}

func TestDB_Find(t *testing.T) {
	_, err := db.Truncate(TestTable)
	assert.NoError(t, err)

	id, err := db.Table(TestTable).InsertGetId(dataMap)
	assert.NoError(t, err)

	res, err := db.Table(TestTable).Find(id)
	assert.NoError(t, err)
	assert.Equal(t, res["id"], int64(id))

	_, err = db.Truncate(TestTable)
	assert.NoError(t, err)
}

func TestDB_WhereExists(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	res, er := db.Table(UsersTable).Select("name").WhereExists(
		db.Table(UsersTable).Select("name").Where("points", ">=", int64(12345)),
	).First()
	assert.NoError(t, er)
	assert.Equal(t, TestUserName, res["name"])

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_WhereNotExists(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	res, er := db.Table(UsersTable).Select("name").WhereNotExists(
		db.Table(UsersTable).Select("name").Where("points", ">=", int64(12345)),
	).First()
	assert.NoError(t, er)
	assert.Equal(t, TestUserName, res["name"])

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_Value(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)
	res, err := db.Table(UsersTable).OrderBy("points", "desc").Value("name")
	assert.NoError(t, err)
	assert.Equal(t, TestUserName, res)

	_, err = db.Table(UsersTable).OrderBy("poin", "desc").Value("name")
	assert.Error(t, err)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_Pluck(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)
	res, err := db.Table(UsersTable).Pluck("name")
	assert.NoError(t, err)

	for k, v := range res {
		resVal := v.(string)
		assert.Equal(t, batchUsers[k]["name"], resVal)
	}

	_, err = db.Table("nonexistent").Pluck("name")
	assert.Error(t, err)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_PluckMap(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)
	res, err := db.Table(UsersTable).PluckMap("name", "points")
	assert.NoError(t, err)

	for k, m := range res {
		for key, value := range m {
			keyVal := key.(string)
			valueVal := value.(int64)
			assert.Equal(t, batchUsers[k]["name"], keyVal)
			assert.Equal(t, batchUsers[k]["points"], valueVal)
		}
	}

	_, err = db.Table("nonexistent").PluckMap("name", "points")
	assert.Error(t, err)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_Exists(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	prepared := db.Table(UsersTable).Select("name").Where("points", ">=", int64(12345))

	exists, err := prepared.Exists()
	assert.NoError(t, err)

	doesntEx, err := prepared.DoesntExists()
	assert.NoError(t, err)

	assert.True(t, exists, "The record must exist at this state of db data")
	assert.False(t, doesntEx, "The record must exist at this state of db data")

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_Count(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	cnt, err := db.Table(UsersTable).Count()
	assert.NoError(t, err)

	assert.Equalf(t, int64(len(batchUsers)), cnt, "want: %d, got: %d", len(batchUsers), cnt)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_Avg(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	avg, err := db.Table(UsersTable).Avg("points")
	assert.NoError(t, err)

	var cntBatch float64
	for _, v := range batchUsers {
		cntBatch += float64(v["points"].(int64)) / float64(len(batchUsers))
	}

	assert.Equalf(t, cntBatch, avg, "want: %d, got: %d", cntBatch, avg)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_MinMax(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	mn, err := db.Table(UsersTable).Min("points")
	assert.NoError(t, err)

	mx, err := db.Table(UsersTable).Max("points")
	assert.NoError(t, err)

	var max float64
	var min = float64(123456)
	for _, v := range batchUsers {
		val := float64(v["points"].(int64))
		if val > max {
			max = val
		}
		if val < min {
			min = val
		}
	}

	assert.Equalf(t, mn, min, "want: %d, got: %d", mn, min)
	assert.Equalf(t, mx, max, "want: %d, got: %d", mx, max)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_Sum(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	sum, err := db.Table(UsersTable).Sum("points")
	assert.NoError(t, err)

	var cntBatch float64
	for _, v := range batchUsers {
		cntBatch += float64(v["points"].(int64))
	}

	assert.Equalf(t, cntBatch, sum, "want: %d, got: %d", cntBatch, sum)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_GroupByHaving(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	res, err := db.Table(UsersTable).Select("points").GroupBy("points").Having("points", ">=", 123).Get()
	assert.NoError(t, err)
	assert.Equal(t, len(res), len(batchUsers)-1)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_HavingRaw(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	res, err := db.Table(UsersTable).Select("points").GroupBy("points").HavingRaw("points > 123").AndHavingRaw("points < 12345").OrHavingRaw("points = 0").Get()
	assert.NoError(t, err)
	assert.Equal(t, len(batchUsers)-3, len(res))

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_AllJoins(t *testing.T) {
	_, err := db.Truncate(PostsTable)
	assert.NoError(t, err)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	err = db.Table(PostsTable).InsertBatch(batchPosts)
	assert.NoError(t, err)

	res, err := db.Table(UsersTable).Select("name", "post", "user_id").InnerJoin(PostsTable, "users.id", "=", "posts.user_id").Get()
	assert.NoError(t, err)

	assert.Equal(t, len(res), len(batchPosts)-1)

	res, err = db.Table(PostsTable).Select("name", "post", "user_id").RightJoin(UsersTable, "posts.user_id", "=", "users.id").Get()
	assert.NoError(t, err)

	assert.Equal(t, len(res), len(batchUsers))

	res, err = db.Table(UsersTable).Select("name", "post", "user_id").FullJoin(PostsTable, "users.id", "=", "posts.user_id").Get()
	assert.NoError(t, err)

	assert.Equal(t, len(res), len(batchUsers)+1)

	res, err = db.Table(UsersTable).Select("name", "post", "user_id").FullJoin(PostsTable, "users.id", "=", "posts.user_id").Get()
	assert.NoError(t, err)

	assert.Equal(t, len(res), len(batchUsers)+1)

	// note InRandomOrder check
	res, err = db.Table(UsersTable).Select("name", "post", "user_id").FullJoin(PostsTable, "users.id", "=", "posts.user_id").InRandomOrder().Get()
	assert.NoError(t, err)

	assert.Equal(t, len(res), len(batchUsers)+1)

	_, err = db.Truncate(PostsTable)
	assert.NoError(t, err)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_OrderByRaw(t *testing.T) {
	_, err := db.Truncate(PostsTable)
	assert.NoError(t, err)

	err = db.Table(PostsTable).InsertBatch(batchPosts)
	assert.NoError(t, err)

	res, err := db.Table(PostsTable).Select("title").OrderByRaw("updated_at - created_at DESC").First()
	assert.NoError(t, err)

	assert.Equal(t, batchPosts[2]["title"], res["title"])

	_, err = db.Truncate(PostsTable)
	assert.NoError(t, err)
}

func TestDB_SelectRaw(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	res, err := db.Table(UsersTable).SelectRaw("SUM(points) as pts").First()
	assert.NoError(t, err)

	var sum int64
	for _, v := range batchUsers {
		sum += v["points"].(int64)
	}
	assert.Equal(t, sum, res["pts"])

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_AndWhereBetween(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	res, err := db.Table(UsersTable).Select("name").WhereBetween("points", 1233, 12345).OrWhereBetween("points", 123456, 67891023).AndWhereNotBetween("points", 12, 23).First()
	assert.NoError(t, err)

	assert.Equal(t, "Darth Vader", res["name"])

	res, err = db.Table(UsersTable).Select("name").WhereNotBetween("points", 12, 123).AndWhereBetween("points", 1233, 12345).OrWhereNotBetween("points", 12, 23).First()
	assert.NoError(t, err)

	assert.Equal(t, "Alex Shmidt", res["name"])

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_WhereRaw(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	res, err := db.Table(UsersTable).Select("name").WhereRaw("LENGTH(name) > 15").OrWhereRaw("points > 1234").Get()
	assert.NoError(t, err)
	assert.Equal(t, len(res), 2)

	cnt, err := db.Table(UsersTable).WhereRaw("points > 123").AndWhereRaw("points < 12345").Count()
	assert.NoError(t, err)
	assert.Equal(t, cnt, int64(1))

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_Offset(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	res, err := db.Table(UsersTable).Offset(2).Limit(10).Get()
	assert.NoError(t, err)
	assert.Equal(t, len(res), 2)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_Rename(t *testing.T) {
	tbl := "tbl1"
	tbl2 := "tbl2"
	_, err := db.DropIfExists(tbl, tbl2)
	assert.NoError(t, err)

	_, err = db.Schema(tbl, func(table *Table) error {
		table.Increments("id")

		return nil
	})
	assert.NoError(t, err)

	_, err = db.Rename(tbl, tbl2)
	assert.NoError(t, err)

	exists, err := db.HasTable("public", tbl2)
	assert.NoError(t, err)
	assert.True(t, exists)

	_, err = db.Drop(tbl2)
	assert.NoError(t, err)
}

func TestDB_WhereIn(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)
	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	res, err := db.Table(UsersTable).Select("name").WhereIn("points", []int64{123, 1234}).OrWhereIn("id", []int64{1, 2}).Get()
	assert.NoError(t, err)
	assert.Equal(t, len(res), 2)

	res, err = db.Table(UsersTable).Select("name").WhereIn("points", []int64{123, 1234}).AndWhereIn("id", []int64{1, 2}).Get()
	assert.NoError(t, err)
	assert.Equal(t, len(res), 2)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_WhereNotIn(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)
	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	res, err := db.Table(UsersTable).Select("name").WhereNotIn("points", []int64{123, 1234}).OrWhereNotIn("id", []int64{1, 2}).Get()
	assert.NoError(t, err)
	assert.Equal(t, len(res), 2)

	res, err = db.Table(UsersTable).Select("name").WhereNotIn("points", []int64{123, 1234}).AndWhereNotIn("id", []int64{1, 2}).Get()
	assert.NoError(t, err)
	assert.Equal(t, len(res), 2)
	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_WhereNull(t *testing.T) {
	_, err := db.Truncate(PostsTable)
	assert.NoError(t, err)

	err = db.Table(PostsTable).InsertBatch(batchPosts)
	assert.NoError(t, err)

	res, err := db.Table(PostsTable).Select("title").WhereNull("post").AndWhereNull("user_id").Get()
	db.Dump()
	assert.NoError(t, err)
	assert.Equal(t, len(res), 1)

	res, err = db.Table(PostsTable).Select("title").WhereNull("post").OrWhereNull("user_id").Get()
	assert.NoError(t, err)
	assert.Equal(t, len(res), 1)

	_, err = db.Truncate(PostsTable)
	assert.NoError(t, err)
}

func TestDB_WhereNotNull(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	res, err := db.Table(UsersTable).Select("name").WhereNotNull("points").AndWhereNotNull("name").Get()
	assert.NoError(t, err)
	assert.Equal(t, len(res), len(batchUsers))

	res, err = db.Table(UsersTable).Select("name").WhereNotNull("points").OrWhereNotNull("name").Get()
	assert.NoError(t, err)
	assert.Equal(t, len(res), len(batchUsers))

	res, err = db.Table(UsersTable).Select("name").Where("id", "=", 1).
		OrWhere("id", "=", 2).AndWhereNotNull("points").Get()
	assert.NoError(t, err)
	assert.Equal(t, len(res), 2)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_LockForUpdate(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	res, err := db.Table(UsersTable).Select("name").LockForUpdate().Get()
	assert.NoError(t, err)
	assert.Equal(t, len(res), len(batchUsers))

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_UnionAll(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	res, err := db.Table(UsersTable).Select("name").UnionAll().Table(UsersTable).Select("name").Get()
	assert.NoError(t, err)
	assert.Equal(t, len(res), len(batchUsers)*2)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_FullOuterJoin(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	res, err := db.Table(UsersTable).Select("name").FullOuterJoin(PostsTable, "users.id", "=", "posts.user_id").Get()
	assert.NoError(t, err)
	assert.Equal(t, len(res), len(batchUsers))

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_Chunk(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)
	var sumOfPoints int64
	err = db.Table(UsersTable).Select("name", "points").Chunk(2, func(users []map[string]interface{}) bool {
		for _, m := range users {
			if val, ok := m["points"]; ok {
				sumOfPoints += val.(int64)
			}
		}
		return true
	})

	assert.NoError(t, err)
	var initialSum int64
	for _, mm := range batchUsers {
		if val, ok := mm["points"]; ok {
			initialSum += val.(int64)
		}
	}
	assert.Equal(t, sumOfPoints, initialSum)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_ChunkFalse(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)
	var sumOfPoints int64
	err = db.Table(UsersTable).Select("name", "points").Chunk(2, func(users []map[string]interface{}) bool {
		for _, m := range users {
			if sumOfPoints > 0 {
				return false
			}
			if val, ok := m["points"]; ok {
				sumOfPoints += val.(int64)
			}
		}
		return true
	})

	assert.NoError(t, err)
	assert.Equal(t, sumOfPoints, batchUsers[0]["points"].(int64))

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_ChunkLessThenAmount(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	var sumOfPoints int64
	err = db.Table(UsersTable).Select("name", "points").Chunk(int64(len(batchUsers)+1), func(users []map[string]interface{}) bool {
		for _, m := range users {
			if val, ok := m["points"]; ok {
				sumOfPoints += val.(int64)
			}
		}
		return true
	})
	assert.NoError(t, err)
	assert.Greater(t, sumOfPoints, int64(0))

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_ChunkLessThenZeroErr(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	err = db.Table(UsersTable).InsertBatch(batchUsers)
	assert.NoError(t, err)

	var sumOfPoints int64
	err = db.Table(UsersTable).Select("name", "points").Chunk(int64(-1), func(users []map[string]interface{}) bool {
		for _, m := range users {
			if val, ok := m["points"]; ok {
				sumOfPoints += val.(int64)
			}
		}
		return true
	})
	assert.Errorf(t, err, "chunk can't be <= 0, your chunk is: -1")

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_ChunkBuilderTableErr(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)
	// reset prev set up table as we don't want to use Table to produce err
	db.Builder.table = ""
	err = db.InsertBatch(batchUsers)
	assert.Error(t, err, errTableCallBeforeOp)

	_, err = db.Select("foo", "bar", "baz").Get()
	assert.Error(t, err, errTableCallBeforeOp)

	err = db.Insert(dataMap)
	assert.Error(t, err, errTableCallBeforeOp)

	_, err = db.InsertGetId(dataMap)
	assert.Error(t, err, errTableCallBeforeOp)

	_, err = db.Update(dataMap)
	assert.Error(t, err, errTableCallBeforeOp)

	_, err = db.Delete()
	assert.Error(t, err, errTableCallBeforeOp)

	_, err = db.Replace(dataMap, "id")
	assert.Error(t, err, errTableCallBeforeOp)

	_, err = db.Increment("clmn", 123)
	assert.Error(t, err, errTableCallBeforeOp)

	_, err = db.Exists()
	assert.Error(t, err, errTableCallBeforeOp)

	_, err = db.Table("nonexistent").Update(dataMap)
	assert.Error(t, err)

	_, err = db.Table("nonexistent").Delete()
	assert.Error(t, err)

	_, err = db.Table("nonexistent").Increment("clmn", 123)
	assert.Error(t, err)

	_, err = db.Table("nonexistent").Replace(dataMap, "id")
	assert.Error(t, err)

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}

func TestDB_FirsNoRecordsErr(t *testing.T) {
	_, err := db.Truncate(UsersTable)
	assert.NoError(t, err)

	_, err = db.Table(TestTable).Select("baz").OrderBy("baz", "desc").OrderBy("foo", "desc").First()
	assert.Errorf(t, err, "no records were produced by query: %s")

	_, err = db.Truncate(UsersTable)
	assert.NoError(t, err)
}
