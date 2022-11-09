// import { ProColumns } from '@ant-design/pro-components';
import { Button, Card, Form, Input, Typography, Table, Select, DatePicker, message, Checkbox, Switch, Spin } from "antd";
import { useAntdTable, useRequest } from 'ahooks';
import moment, { Moment } from "moment";
import { useEffect, useState } from "react";

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
    responsive: ['sm'],
  },
  {
    title: 'Message',
    dataIndex: 'Message',
  },
  {
    title: 'User',
    dataIndex: 'User',
    width: 100,
    responsive: ['lg'],
  },
  {
    title: 'TraceNo',
    dataIndex: 'TraceNo',
    align: "center",
    width: 150,
    responsive: ['xl'],
  },
  {
    title: 'Date',
    align: "center",
    width: 170,
    render: (_: any, x: LogEntry) => moment(x.CreatedOnUtc).local().format("MM/DD/YYYY hh:mm:ss A"),
    responsive: ['md'],
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
  // console.log(formData);

  var body = {
    PageSize: pageSize,
    PageIndex: current,
    DBName: formData.DBName,
    TableName: formData.TableName,
    User: formData.User?.trim(),
    TraceNo: formData.TraceNo?.trim(),
    Message: formData.Message?.trim(),
    StartTime: "",
    EndTime: "",
    Level: formData.Level,
    Flags: 0,
  };

  // date range
  var dateRange: Moment[] = formData.DateRange;
  if (dateRange?.length == 2) {
    var start = dateRange[0];
    var end = dateRange[1];
    if (start) {
      var startTime = moment(start);  // Must clone
      body.StartTime = startTime.startOf('day').utc().format();
    }
    if (end) {
      var endTime = moment(end);      // Must clone
      body.EndTime = endTime.startOf('day').utc().format();
    }
  }
  // flags, fuzzy search
  if (formData.FuzzySearch) {
    body.Flags |= 1;
  }

  console.log(body);

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

const getClients = async () => {
  var url = "http://localhost:7160/api/clients"
  // console.log(formData);

  var body = {
  };

  const resp = await fetch(url, {
    method: "POST",
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  });
  const jsonStr = await resp.text();
  const r = JSON.parse(jsonStr);
  // console.log(r);

  return r;
};

const getDatabases = async (clientID: string) => {
  var url = "http://localhost:7160/api/dbs"

  var body = {
    ClientID: clientID
  };

  const resp = await fetch(url, {
    method: "POST",
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  });
  const jsonStr = await resp.text();
  const r = JSON.parse(jsonStr);
  // console.log(r);

  return r;
};

const getTables = async (db: string) => {
  var url = "http://localhost:7160/api/tables"

  var body = {
    Database: db,
  };

  const resp = await fetch(url, {
    method: "POST",
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  });
  const jsonStr = await resp.text();
  const r = JSON.parse(jsonStr);
  // console.log(r);

  return r;
};

const tableDom = function (clients: string[]) {
  const [form] = Form.useForm();

  const [dbs, setDBs] = useState([]);
  const [tables, setTables] = useState([]);

  const { tableProps, search } = useAntdTable(getTableData, {
    defaultPageSize: 10,
    form,
    manual: true,
    // defaultParams: [
    //   { current: 1, pageSize: 10 },
    //   {
    //     Level: -1,
    //     Client: window.localStorage.getItem('client'),
    //     DBName: window.localStorage.getItem('db'),
    //     TableName: window.localStorage.getItem('table'),
    //     DateRange: [moment().subtract(1, 'M'), null],
    //   },
    // ],
  });

  var client = window.localStorage.getItem('client');
  var db = window.localStorage.getItem('db');
  var table = window.localStorage.getItem('table');

  return <div>
    <Form form={form} layout="inline" size="small"
      initialValues={{
        Level: -1,
        Client: client,
        DBName: db,
        TableName: table,
        DateRange: [moment().subtract(1, 'M'), null],
      }}
    >
      <Form.Item name="Client" rules={[{ required: true, message: 'Client is required' }]}>
        <Select placeholder="Client"
          dropdownMatchSelectWidth={false}
          onChange={async (clientID) => {
            var r = await getDatabases(clientID);
            window.localStorage.setItem('client', clientID);
            window.localStorage.removeItem('db');
            window.localStorage.removeItem('table');
            setDBs(r?.Databases ?? []);
          }}
          options={clients?.map(x => ({ label: x, value: x }))}
        ></Select>
      </Form.Item>
      <Form.Item name="DBName" rules={[{ required: true, message: 'Database is required' }]}>
        <Select placeholder="Database"
          dropdownMatchSelectWidth={false}
          onChange={async (db) => {
            var r = await getTables(db);
            window.localStorage.setItem('db', db);
            window.localStorage.removeItem('table');
            setTables(r?.Tables ?? []);
          }}
          options={dbs?.map(x => ({ label: x, value: x }))}
        ></Select>
      </Form.Item>
      <Form.Item name="TableName" rules={[{ required: true, message: 'Table is required' }]}>
        <Select placeholder="Table"
          dropdownMatchSelectWidth={false}
          onChange={async (table) => {
            window.localStorage.setItem('table', table);
          }}
          options={tables?.map(x => ({ label: x, value: x }))}
        ></Select>
      </Form.Item>
      <Form.Item name="Level">
        <Select placeholder="Level" dropdownMatchSelectWidth={false}>
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
      <Form.Item name="DateRange" rules={[{
        // validator: (_, value) => value ? Promise.resolve() : Promise.reject(new Error('Should accept agreement')),
        validator: (_, array) => {
          if (!array || !array[0]) {
            return Promise.reject(new Error('Start date is required'));
          }

          return Promise.resolve();
        }
      }]}>
        <RangePicker allowEmpty={[true, true]} />
      </Form.Item>
      <Form.Item name="FuzzySearch" valuePropName="checked">
        <Switch checkedChildren="Fuzzy" unCheckedChildren="Fuzzy" />
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
      columns={columns}
      scroll={{
        x: true,
      }}
      {...tableProps}
    />
  </div >;
};

export default function HomePage() {
  const { data } = useRequest(getClients);
  // console.log(data);

  return tableDom(data?.LogClients);
}
