package serializer

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/mi991203/TestMoon/backend/testing/execution/model"
	"github.com/mi991203/TestMoon/libs/go/transport"
	pb "github.com/mi991203/TestMoon/protocols/testing/execution/v1"
)

func ListTasks(value any) (*pb.ListTasksResponse, error) {
	out := &pb.ListTasksResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func CreateTask(value any) (*pb.CreateTaskResponse, error) {
	out := &pb.CreateTaskResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func UpdateTask(value any) (*pb.UpdateTaskResponse, error) {
	out := &pb.UpdateTaskResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func DeleteTask(value any) (*pb.DeleteTaskResponse, error) {
	out := &pb.DeleteTaskResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func ListRuns(value any) (*pb.ListRunsResponse, error) {
	out := &pb.ListRunsResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func GetRun(value any) (*pb.GetRunResponse, error) {
	out := &pb.GetRunResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func CancelRun(value any) (*pb.CancelRunResponse, error) {
	out := &pb.CancelRunResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func ListProjectRunners(value any) (*pb.ListProjectRunnersResponse, error) {
	out := &pb.ListProjectRunnersResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func RevokePrivateRunner(value any) (*pb.RevokePrivateRunnerResponse, error) {
	out := &pb.RevokePrivateRunnerResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func ListPublicRunners(value any) (*pb.ListPublicRunnersResponse, error) {
	out := &pb.ListPublicRunnersResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func RevokePublicRunner(value any) (*pb.RevokePublicRunnerResponse, error) {
	out := &pb.RevokePublicRunnerResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func RegisterRunner(value any) (*pb.RegisterRunnerResponse, error) {
	out := &pb.RegisterRunnerResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func RunnerHeartbeat(value any) (*pb.RunnerHeartbeatResponse, error) {
	out := &pb.RunnerHeartbeatResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func PollRunnerJob(value any) (*pb.PollRunnerJobResponse, error) {
	out := &pb.PollRunnerJobResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func ProgressRun(value any) (*pb.ProgressRunResponse, error) {
	out := &pb.ProgressRunResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func CompleteRun(value any) (*pb.CompleteRunResponse, error) {
	out := &pb.CompleteRunResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func optionalTime(value *time.Time) *timestamppb.Timestamp {
	if value == nil {
		return nil
	}
	return timestamppb.New(*value)
}

func Run(value model.Run) *pb.Run {
	return &pb.Run{Id: value.ID, OrgId: value.OrgID, ProjectId: value.ProjectID, Name: value.Name, CreatedAt: timestamppb.New(value.CreatedAt), UpdatedAt: timestamppb.New(value.UpdatedAt), UserId: value.UserID, TaskId: value.TaskID, State: value.State, TargetRunnerId: value.TargetRunnerID, NodeId: value.NodeID, TimeoutSeconds: int32(value.TimeoutSeconds), StartedAt: optionalTime(value.StartedAt), FinishedAt: optionalTime(value.FinishedAt), Error: value.Error}
}

func Runner(value model.Runner) *pb.Runner {
	return &pb.Runner{Id: value.ID, OrgId: value.OrgID, ProjectId: value.ProjectID, Name: value.Name, CreatedAt: timestamppb.New(value.CreatedAt), UpdatedAt: timestamppb.New(value.UpdatedAt), Public: value.Public, Revoked: value.Revoked, Registered: value.Registered, Concurrency: int32(value.Concurrency), Version: value.Version, LastSeen: optionalTime(value.LastSeen)}
}
