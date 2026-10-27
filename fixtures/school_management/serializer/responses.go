package serializer

import (
	"github.com/mi991203/TestMoon/fixtures/school_management/model"
	pb "github.com/mi991203/TestMoon/protocols/fixtures/school_management/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func School(v model.School) *pb.School {
	return &pb.School{Id: v.ID, CreatedAt: timestamppb.New(v.CreatedAt), UpdatedAt: timestamppb.New(v.UpdatedAt),
		Name:    v.Name,
		Address: v.Address,
	}
}

func Class(v model.Class) *pb.Class {
	return &pb.Class{Id: v.ID, CreatedAt: timestamppb.New(v.CreatedAt), UpdatedAt: timestamppb.New(v.UpdatedAt),
		SchoolId: v.SchoolID,
		Name:     v.Name,
	}
}

func Student(v model.Student) *pb.Student {
	return &pb.Student{Id: v.ID, CreatedAt: timestamppb.New(v.CreatedAt), UpdatedAt: timestamppb.New(v.UpdatedAt),
		ClassId:   v.ClassID,
		Name:      v.Name,
		StudentNo: v.StudentNo,
	}
}
