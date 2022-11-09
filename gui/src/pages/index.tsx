// import { ProColumns } from '@ant-design/pro-components';
import { Button, Card, Form, Input, Typography, Table, Select, DatePicker } from "antd";
import { useAntdTable } from 'ahooks';
import moment, { Moment } from "moment";

const { Text } = Typography;
const { RangePicker } = DatePicker;

const columns: any = [
  {
    title: 'Level',
    render: (_: any, x: LogEntry) => {
      switch (x.Level) {
        case 0:
          return <Text type="secondary">Verbose</Text>;
        case 1:
          return <Text type="secondary">Debug</Text>;
        case 2:
          return <Text>Info</Text>;
        case 3:
          return <Text type="warning">Warning</Text>;
        case 4:
          return <Text type="danger">Error</Text>;
        case 5:
          return <Text type="danger">Fatal</Text>;
        default:
          return <Text type="secondary">Unknown</Text>;
      }
    },
    align: "center",
    width: 60,
  },
  {
    title: 'Message',
    dataIndex: 'Message',
  },
  {
    title: 'User',
    dataIndex: 'User',
    width: 100,
  },
  {
    title: 'TraceNo',
    dataIndex: 'TraceNo',
    align: "center",
    width: 150,
  },
  {
    title: 'Date',
    align: "center",
    width: 170,
    render: (_: any, x: LogEntry) => moment(x.CreatedOnUtc).local().format("MM/DD/YYYY hh:mm:ss A"),
  },
];

interface LogEntry {
  ID: string,
  TraceNo: string,
  User: string,
  Message: string,
  Error: string,
  StackTrace: string,
  Level: number,
  CreatedOnUtc: number,
}

const getTableData = async ({ current, pageSize }: any, formData: any) => {
  var url = "http://localhost:7160/api/logs"
  // var url = "http://localhost:7160/api/logs?DBName=LOG_DL&TableName=2022&PageSize=" + pageSize + "&PageIndex=" + current
  console.log(formData);

  var body = {
    PageSize: pageSize,
    PageIndex: current,
    DBName: formData.DBName,
    TableName: formData.TableName,
    Level: formData.Level,
    User: formData.User,
    TraceNo: formData.TraceNo,
    Message: formData.Message,
    StartTime: 0,
    EndTime: 0,
  };

  var dateRange: Moment[] = formData.DateRange;
  if (dateRange?.length == 2) {
    var start = dateRange[0];
    var end = dateRange[1];
    if (start) {
      body.StartTime = moment(start.local().format("YYYY-MM-DD")).valueOf()
    }
    if (end) {
      body.EndTime = moment(end.local().format("YYYY-MM-DD")).valueOf()
    }
  }

  const resp = await fetch(url, {
    method: "POST",
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  });
  const jsonStr = await resp.text();
  const r = JSON.parse(jsonStr);

  return {
    list: r.LogEntries,
    total: r.TotalCount,
  };
};

export default function HomePage() {
  const [form] = Form.useForm();

  const { tableProps, search } = useAntdTable(getTableData, {
    defaultPageSize: 10,
    form,
    manual: true,
  });

  return <Card>
    <Form form={form} layout="inline" size="small" initialValues={{ Level: -1, DBName: "DL", TableName: "2022" }}>
      <Form.Item name="DBName">
        <Select placeholder="Client">
          <Select.Option value="DL">DL</Select.Option>
        </Select>
      </Form.Item>
      <Form.Item name="TableName">
        <Select placeholder="Archive">
          <Select.Option value="2022">2022</Select.Option>
        </Select>
      </Form.Item>
      <Form.Item name="Level">
        <Select placeholder="Level">
          <Select.Option value={-1}>All</Select.Option>
          <Select.Option value={0}>Verbose</Select.Option>
          <Select.Option value={1}>Debug</Select.Option>
          <Select.Option value={2}>Info</Select.Option>
          <Select.Option value={3}>Warning</Select.Option>
          <Select.Option value={4}>Error</Select.Option>
          <Select.Option value={5}>Fatal</Select.Option>
        </Select>
      </Form.Item>
      <Form.Item name="Message">
        <Input placeholder="Message" />
      </Form.Item>
      <Form.Item name="User">
        <Input placeholder="User" />
      </Form.Item>
      <Form.Item name="TraceNo">
        <Input placeholder="TraceNo" />
      </Form.Item>
      <Form.Item name="DateRange">
        <RangePicker allowEmpty={[true, true]} />
      </Form.Item>
      <Form.Item>
        <Button
          type="primary"
          htmlType="submit"
          style={{ marginRight: 20 }}
          onClick={search.submit}
        >
          Search
        </Button>
        <Button onClick={search.reset}>Reset</Button>
      </Form.Item>
    </Form>
    <br />
    <Table rowKey="ID"
      size="small"
      // dataSource={LogEntries}
      columns={columns}
      {...tableProps}
    />
  </Card >;
}
