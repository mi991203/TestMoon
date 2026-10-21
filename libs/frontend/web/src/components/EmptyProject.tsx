import { Alert, Card, message } from "antd";

export function EmptyProject() {
  return (
    <Card>
      <Alert
        type="info"
        showIcon
        message="先选择组织和项目"
        description="组织管理员可以创建项目，并为组织成员授予项目角色。"
      />
    </Card>
  );
}
