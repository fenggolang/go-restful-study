package main

import (
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"net/http"
	"log"
	"github.com/go-openapi/spec"
)

// User是一个简单的类型
type User struct {
	ID string `json:"id" description:"identifier of the user"`
	Name string `json:"name" description:"name of the user" default:"john"`
	Age int `json:"age" description:"age of the user" default:"21"`
}

// UserResource是User domain的Rest层
type UserResource struct {
	// 通常会使用DAO(数据访问对象)
	users map[string]User
}

// GET http://localhost:8080/users
func (u UserResource) findAllUsers(request *restful.Request,response *restful.Response){
	list:=[]User{}
	for _,each:=range u.users{
		list = append(list,each)
	}
	response.WriteEntity(list)
}

// GET http://localhost:8080/users/1
func (u UserResource) findUser(request *restful.Request,response *restful.Response){
	id:=request.PathParameter("user-id")
	usr:=u.users[id]
	if len(usr.ID) == 0{
		response.WriteErrorString(http.StatusNotFound,"User could not be found.")
	} else{
		response.WriteEntity(usr)
	}
}
// PUT http://localhost:8080/users/1
// <User><Id>1</Id><Name>Melissa Raspherry</Name></User>
func (u *UserResource) updateUser(request *restful.Request,response *restful.Response){
	usr:=new(User)
	err:=request.ReadEntity(&usr)
	if err == nil{
		u.users[usr.ID] = *usr
		response.WriteEntity(usr)
	}else {
		response.WriteError(http.StatusNotFound,err)
	}
}

// POST http://localhost:8080/users/1
// <User><Id>1</Id><Name>Melissa</Name></User>
func (u *UserResource) createUser(request *restful.Request,response *restful.Response){
	usr:=User{ID:request.PathParameter("user-id")}
	err:=request.ReadEntity(&usr)
	if err == nil{
		u.users[usr.ID] = usr
		response.WriteHeaderAndEntity(http.StatusCreated,usr)
	}else{
		response.WriteError(http.StatusInternalServerError,err)
	}
}

// DELETE http://localhost:8080/users/1
func (u *UserResource) removeUser(request *restful.Request,response *restful.Response){
	id:=request.PathParameter("user-id")
	delete(u.users,id)
}

func enrichSwaggerObject(swo *spec.Swagger){
	swo.Info = &spec.Info{
		InfoProps:spec.InfoProps{
			Title:"UserService",
			Description:"Resource for managing Users",
			Contact:&spec.ContactInfo{
				Name:"john",
				Email:"john@doe.rp",
				URL:"http://johndoe.org",
			},
			License:&spec.License{
				Name:"MIT",
				URL:"http://mit.org",
			},
			Version:"1.0.0",
		},
	}
	swo.Tags = []spec.Tag{spec.Tag{TagProps:spec.TagProps{
		Name:"users",
		Description:"Managing users",
	}}}
}

// WebService 创建一个可以处理用户资源的REST请求的新服务
func (u UserResource) WebService() *restful.WebService{
	ws:=new(restful.WebService)
	ws.Path("/users").
		Consumes(restful.MIME_XML,restful.MIME_JSON).
		Produces(restful.MIME_JSON,restful.MIME_XML) // 也可以指定每条路线

	tags:=[]string{"users"}

	ws.Route(ws.GET("/").To(u.findAllUsers).
		// docs
		Doc("get all users").
		Metadata(restfulspec.KeyOpenAPITags,tags).
		Writes([]User{}).
		Returns(200,"OK",[]User{}))
	ws.Route(ws.GET("/{user-id}").To(u.findUser).
		// docs
		Doc("get a user").
		Param(ws.PathParameter("user-id","identifier of the user").DataType("integer").DefaultValue("1")).
		Metadata(restfulspec.KeyOpenAPITags,tags).
		Writes(User{}). // on the response
		Returns(200,"OK",User{}).
		Returns(404,"Not Found",nil))
	ws.Route(ws.PUT("/{user-id}").To(u.updateUser).
		//docs
		Doc("update a user").
		Param(ws.PathParameter("user-id","identifier of the user").DataType("string")).
		Metadata(restfulspec.KeyOpenAPITags,tags).
		Reads(User{}), // from the request
		)
	ws.Route(ws.POST("").To(u.createUser).
		// docs
		Doc("create a user").
		Metadata(restfulspec.KeyOpenAPITags,tags).
		Reads(User{})) // from the request
	ws.Route(ws.DELETE("/{user-id}").To(u.removeUser).
		// docs
		Doc("delete a user").
		Metadata(restfulspec.KeyOpenAPITags,tags).
		Param(ws.PathParameter("user-id","identifier of the user").DataType("string")))

	return ws
}

func main() {
	u:=UserResource{map[string]User{}}
	restful.DefaultContainer.Add(u.WebService())

	config:=restfulspec.Config{
		WebServices:restful.RegisteredWebServices(),
		APIPath:"/apidocs.json",
		PostBuildSwaggerObjectHandler:enrichSwaggerObject,
	}
	restful.DefaultContainer.Add(restfulspec.NewOpenAPIService(config))

	// (可选)可以安装Swagger服务，它在REST API上提供了一个很好的Web UI
	// 你需要下载Swagger HTML5资源并在下面的配置中更改FilePath位置
	// 打开 http://localhost:8080/apidocs/?url=http://localhost:8080/apidocs.json
	http.Handle("/apidocs/",http.StripPrefix("/apidocs/",http.FileServer(http.Dir("/Users/emicklei/Projects/swagger-ui/dist"))))

	log.Printf("start listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080",nil))
}

// 测试
// GET请求查询：curl -XGET http://localhost:8080/users/1
// POST插入数据：curl -XPOST http://localhost:8080/users/1