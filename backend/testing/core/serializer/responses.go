package serializer

import (
	"github.com/mi991203/TestMoon/libs/go/transport"
	pb "github.com/mi991203/TestMoon/protocols/testing/core/v1"
)

func ListAPIs(value any) (*pb.ListAPIsResponse, error) {
	out := &pb.ListAPIsResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func GetAPI(value any) (*pb.GetAPIResponse, error) {
	out := &pb.GetAPIResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func CreateAPI(value any) (*pb.CreateAPIResponse, error) {
	out := &pb.CreateAPIResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func UpdateAPI(value any) (*pb.UpdateAPIResponse, error) {
	out := &pb.UpdateAPIResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func DeleteAPI(value any) (*pb.DeleteAPIResponse, error) {
	out := &pb.DeleteAPIResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func ListTestCases(value any) (*pb.ListTestCasesResponse, error) {
	out := &pb.ListTestCasesResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func GetTestCase(value any) (*pb.GetTestCaseResponse, error) {
	out := &pb.GetTestCaseResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func CreateTestCase(value any) (*pb.CreateTestCaseResponse, error) {
	out := &pb.CreateTestCaseResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func UpdateTestCase(value any) (*pb.UpdateTestCaseResponse, error) {
	out := &pb.UpdateTestCaseResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func DeleteTestCase(value any) (*pb.DeleteTestCaseResponse, error) {
	out := &pb.DeleteTestCaseResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func ListScenarios(value any) (*pb.ListScenariosResponse, error) {
	out := &pb.ListScenariosResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func GetScenario(value any) (*pb.GetScenarioResponse, error) {
	out := &pb.GetScenarioResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func CreateScenario(value any) (*pb.CreateScenarioResponse, error) {
	out := &pb.CreateScenarioResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func UpdateScenario(value any) (*pb.UpdateScenarioResponse, error) {
	out := &pb.UpdateScenarioResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func DeleteScenario(value any) (*pb.DeleteScenarioResponse, error) {
	out := &pb.DeleteScenarioResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func ListEnvironments(value any) (*pb.ListEnvironmentsResponse, error) {
	out := &pb.ListEnvironmentsResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func GetEnvironment(value any) (*pb.GetEnvironmentResponse, error) {
	out := &pb.GetEnvironmentResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func CreateEnvironment(value any) (*pb.CreateEnvironmentResponse, error) {
	out := &pb.CreateEnvironmentResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func UpdateEnvironment(value any) (*pb.UpdateEnvironmentResponse, error) {
	out := &pb.UpdateEnvironmentResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func DeleteEnvironment(value any) (*pb.DeleteEnvironmentResponse, error) {
	out := &pb.DeleteEnvironmentResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func ListFolders(value any) (*pb.ListFoldersResponse, error) {
	out := &pb.ListFoldersResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func GetFolder(value any) (*pb.GetFolderResponse, error) {
	out := &pb.GetFolderResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func CreateFolder(value any) (*pb.CreateFolderResponse, error) {
	out := &pb.CreateFolderResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func UpdateFolder(value any) (*pb.UpdateFolderResponse, error) {
	out := &pb.UpdateFolderResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func DeleteFolder(value any) (*pb.DeleteFolderResponse, error) {
	out := &pb.DeleteFolderResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func ListProjects(value any) (*pb.ListProjectsResponse, error) {
	out := &pb.ListProjectsResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func CreateProject(value any) (*pb.CreateProjectResponse, error) {
	out := &pb.CreateProjectResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func UpdateProject(value any) (*pb.UpdateProjectResponse, error) {
	out := &pb.UpdateProjectResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func DeleteProject(value any) (*pb.DeleteProjectResponse, error) {
	out := &pb.DeleteProjectResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func Authorize(value any) (*pb.AuthorizeResponse, error) {
	out := &pb.AuthorizeResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func ListProjectMembers(value any) (*pb.ListProjectMembersResponse, error) {
	out := &pb.ListProjectMembersResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func SetProjectMember(value any) (*pb.SetProjectMemberResponse, error) {
	out := &pb.SetProjectMemberResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func RemoveProjectMember(value any) (*pb.RemoveProjectMemberResponse, error) {
	out := &pb.RemoveProjectMemberResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func CreateSnapshot(value any) (*pb.CreateSnapshotResponse, error) {
	out := &pb.CreateSnapshotResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func PreviewImport(value any) (*pb.PreviewImportResponse, error) {
	out := &pb.PreviewImportResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func CommitImport(value any) (*pb.CommitImportResponse, error) {
	out := &pb.CommitImportResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func ListShares(value any) (*pb.ListSharesResponse, error) {
	out := &pb.ListSharesResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func CreateShare(value any) (*pb.CreateShareResponse, error) {
	out := &pb.CreateShareResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func DeleteShare(value any) (*pb.DeleteShareResponse, error) {
	out := &pb.DeleteShareResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func ReadSharedDocument(value any) (*pb.ReadSharedDocumentResponse, error) {
	out := &pb.ReadSharedDocumentResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}
