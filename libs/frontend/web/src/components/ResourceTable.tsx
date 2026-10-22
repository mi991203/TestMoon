import { Button, Space, Table } from "antd";

interface Row {
  id: string;
  name?: string;
  revision?: string | number;
  updated_at?: string;
}
export function ResourceTable<T extends Row>({
  items,
  loading,
  onEdit,
  onDelete,
  readOnly = false,
}: {
  items: T[];
  loading?: boolean;
  onEdit: (row: T) => void;
  onDelete?: (row: T) => void;
  readOnly?: boolean;
}) {
  return (
    <Table<T>
      rowKey="id"
      loading={loading}
      dataSource={items}
      pagination={{ pageSize: 15, showSizeChanger: true }}
      columns={[
        {
          title: "名称",
          dataIndex: "name",
          render: (name, row) => <a onClick={() => onEdit(row)}>{name}</a>,
        },
        { title: "版本", dataIndex: "revision", width: 80 },
        {
          title: "更新时间",
          dataIndex: "updated_at",
          render: (v) => (v ? new Date(v).toLocaleString() : "—"),
        },
        {
          title: "操作",
          width: 180,
          render: (_, row) => (
            <Space>
              <Button size="small" onClick={() => onEdit(row)}>
                {readOnly ? "查看" : "编辑"}
              </Button>
              {onDelete && !readOnly && (
                <Button size="small" danger onClick={() => onDelete(row)}>
                  删除
                </Button>
              )}
            </Space>
          ),
        },
      ]}
    />
  );
}
