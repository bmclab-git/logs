// import { ProColumns } from '@ant-design/pro-components';
import { Button, Form, Input, Typography, Table, Select, DatePicker, Switch, Modal, InputNumber } from "antd";
import { useAntdTable } from 'ahooks';
import moment, { Moment } from "moment";
import React, { useEffect, useState } from "react";
import ReactJson from 'react-json-view'

const IS_PRODUCTION = process.env.NODE_ENV === "production";
export const LOCAL_API_ROOT = IS_PRODUCTION ? "/api" : "http://localhost:7160/api";

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
    render: (_: any, x: LogEntry) => <a onClick={() => showDetails(x)}>{x.Message}</a>,
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

const showDetails = (x: LogEntry) => {
  // var json = JSON.stringify(x, null, "  ");
  // json = json.replaceAll("\\n", "\n");
  // json = json.replaceAll("\\t", "\t");
  // const content = <pre className="code">{json}</pre>;
  const content = <ReactJson src={x} theme="monokai" iconStyle="circle" displayDataTypes={false} />;

  Modal.info({
    icon: null,
    closable: true,
    content: content,
    width: "100%",
  });

  // const timer = setInterval(() => {
  //   secondsToGo -= 1;
  //   modal.update({
  //     content: `This modal will be destroyed after ${secondsToGo} second.`,
  //   });
  // }, 1000);

  // setTimeout(() => {
  //   clearInterval(timer);
  //   modal.destroy();
  // }, secondsToGo * 1000);
};

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
  var url = LOCAL_API_ROOT + "/logs"
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

async function getListData(client: string, db: string) {
  var url = LOCAL_API_ROOT + `/listData?client=${client}&db=${db}`

  const resp = await fetch(url);
  const jsonStr = await resp.text();
  const r = JSON.parse(jsonStr);

  return r;
};

const tableDom = function (listData: ListData, setListData: React.Dispatch<React.SetStateAction<ListData>>) {
  const [form] = Form.useForm();

  const { tableProps, search } = useAntdTable(getTableData, {
    defaultPageSize: 10,
    form,
    manual: true,
  });

  return <div>
    <Form form={form} layout="inline" size="small"
      initialValues={{
        Level: -1,
        DateRange: [moment().subtract(1, 'M'), null],
      }}
    >
      <Form.Item name="Client" rules={[{ required: true, message: 'Client is required' }]}>
        <Select placeholder="Client"
          dropdownMatchSelectWidth={false}
          onChange={async (client) => {
            var r = await getListData(client, "");

            form.resetFields(["DBName", "TableName"]);

            setListData({
              ...r,
              Client: client,
              Database: "",
              Table: "",
            });
          }}
          options={listData?.Clients?.map(x => ({ label: x, value: x }))}
        ></Select>
      </Form.Item>
      <Form.Item name="DBName" rules={[{ required: true, message: 'Database is required' }]}>
        <Select placeholder="Database"
          dropdownMatchSelectWidth={false}
          onChange={async (db) => {
            var r = await getListData(listData.Client, db);

            form.resetFields(["TableName"]);

            setListData({
              ...r,
              Database: db,
            });
          }}
          options={listData?.Databases?.map(x => ({ label: x, value: x }))}
        ></Select>
      </Form.Item>
      <Form.Item name="TableName" rules={[{ required: true, message: 'Table is required' }]}>
        <Select placeholder="Table"
          dropdownMatchSelectWidth={false}
          onChange={async (table) => {
            // window.localStorage.setItem('table', table);
            setListData({
              ...listData,
              Table: table,
            });
          }}
          options={listData.Tables?.map(x => ({ label: x, value: x }))}
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
      <Form.Item name="Flags">
        <InputNumber placeholder="Flags" />
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

interface ListData {
  Client: string,
  Database: string,
  Table: string,
  Clients: string[],
  Databases: string[],
  Tables: string[],
}

export default function HomePage() {
  const [listData, setListData] = useState<ListData>({
    Client: "",
    Database: "",
    Table: "",
    Clients: [],
    Databases: [],
    Tables: [],
  });

  useEffect(() => {
    getListData("", "").then(rs => {
      setListData(rs);
    })
  }, []);

  return tableDom(listData, setListData);
}
