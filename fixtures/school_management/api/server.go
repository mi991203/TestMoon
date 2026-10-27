package api

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/mi991203/TestMoon/fixtures/school_management/logic"
	"github.com/mi991203/TestMoon/fixtures/school_management/model"
	"github.com/mi991203/TestMoon/fixtures/school_management/serializer"
	pb "github.com/mi991203/TestMoon/protocols/fixtures/school_management/v1"
)

type Server struct {
	pb.UnimplementedSchoolManagementServer
	Service *logic.Service
}

func (s *Server) CreateSchool(ctx context.Context, req *pb.CreateSchoolRequest) (*pb.School, error) {
	v, err := s.Service.CreateSchool(ctx, model.School{Name: req.GetName(), Address: req.GetAddress()})
	if err != nil {
		return nil, err
	}
	return serializer.School(v), nil
}

func (s *Server) GetSchool(ctx context.Context, req *pb.GetSchoolRequest) (*pb.School, error) {
	v, err := s.Service.GetSchool(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return serializer.School(v), nil
}

func (s *Server) UpdateSchool(ctx context.Context, req *pb.UpdateSchoolRequest) (*pb.School, error) {
	v, err := s.Service.UpdateSchool(ctx, model.School{Record: model.Record{ID: req.GetId()}, Name: req.GetName(), Address: req.GetAddress()})
	if err != nil {
		return nil, err
	}
	return serializer.School(v), nil
}

func (s *Server) DeleteSchool(ctx context.Context, req *pb.DeleteSchoolRequest) (*emptypb.Empty, error) {
	if err := s.Service.DeleteSchool(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListSchools(ctx context.Context, req *pb.ListSchoolsRequest) (*pb.ListSchoolsResponse, error) {
	values, err := s.Service.ListSchools(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*pb.School, 0, len(values))
	for _, v := range values {
		items = append(items, serializer.School(v))
	}
	return &pb.ListSchoolsResponse{Items: items}, nil
}

func (s *Server) CreateClass(ctx context.Context, req *pb.CreateClassRequest) (*pb.Class, error) {
	v, err := s.Service.CreateClass(ctx, model.Class{SchoolID: req.GetSchoolId(), Name: req.GetName()})
	if err != nil {
		return nil, err
	}
	return serializer.Class(v), nil
}

func (s *Server) GetClass(ctx context.Context, req *pb.GetClassRequest) (*pb.Class, error) {
	v, err := s.Service.GetClass(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return serializer.Class(v), nil
}

func (s *Server) UpdateClass(ctx context.Context, req *pb.UpdateClassRequest) (*pb.Class, error) {
	v, err := s.Service.UpdateClass(ctx, model.Class{Record: model.Record{ID: req.GetId()}, SchoolID: req.GetSchoolId(), Name: req.GetName()})
	if err != nil {
		return nil, err
	}
	return serializer.Class(v), nil
}

func (s *Server) DeleteClass(ctx context.Context, req *pb.DeleteClassRequest) (*emptypb.Empty, error) {
	if err := s.Service.DeleteClass(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListClasses(ctx context.Context, req *pb.ListClassesRequest) (*pb.ListClassesResponse, error) {
	values, err := s.Service.ListClasses(ctx, req.GetSchoolId())
	if err != nil {
		return nil, err
	}
	items := make([]*pb.Class, 0, len(values))
	for _, v := range values {
		items = append(items, serializer.Class(v))
	}
	return &pb.ListClassesResponse{Items: items}, nil
}

func (s *Server) CreateStudent(ctx context.Context, req *pb.CreateStudentRequest) (*pb.Student, error) {
	v, err := s.Service.CreateStudent(ctx, model.Student{ClassID: req.GetClassId(), Name: req.GetName(), StudentNo: req.GetStudentNo()})
	if err != nil {
		return nil, err
	}
	return serializer.Student(v), nil
}

func (s *Server) GetStudent(ctx context.Context, req *pb.GetStudentRequest) (*pb.Student, error) {
	v, err := s.Service.GetStudent(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return serializer.Student(v), nil
}

func (s *Server) UpdateStudent(ctx context.Context, req *pb.UpdateStudentRequest) (*pb.Student, error) {
	v, err := s.Service.UpdateStudent(ctx, model.Student{Record: model.Record{ID: req.GetId()}, ClassID: req.GetClassId(), Name: req.GetName(), StudentNo: req.GetStudentNo()})
	if err != nil {
		return nil, err
	}
	return serializer.Student(v), nil
}

func (s *Server) DeleteStudent(ctx context.Context, req *pb.DeleteStudentRequest) (*emptypb.Empty, error) {
	if err := s.Service.DeleteStudent(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListStudents(ctx context.Context, req *pb.ListStudentsRequest) (*pb.ListStudentsResponse, error) {
	values, err := s.Service.ListStudents(ctx, req.GetSchoolId(), req.GetClassId())
	if err != nil {
		return nil, err
	}
	items := make([]*pb.Student, 0, len(values))
	for _, v := range values {
		items = append(items, serializer.Student(v))
	}
	return &pb.ListStudentsResponse{Items: items}, nil
}
