// import { ProColumns } from '@ant-design/pro-components';
import { Spin, Table } from "antd";
import { useRequest } from 'ahooks';
import { useState } from "react";
import moment from "moment";

const columns: any = [
  {
    title: 'Level',
    render: (_: any, x: LogEntry) => {
      switch (x.Level) {
        case 0:
          return "Verbose";
        case 1:
          return "Debug";
        case 2:
          return "Warning";
        case 3:
          return "Error";
        case 4:
          return "Fatal";
        default:
          return "Unknown";
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
    width: 50,
  },
  {
    title: 'TraceNo',
    dataIndex: 'TraceNo',
    align: "center",
    width: 50,
  },
  {
    title: 'Date',
    align: "center",
    width: 170,
    render: (_: any, x: LogEntry) => moment(x.CreatedOnUtc).local().format("MM/DD/YYYY hh:mm:ss A"),
  },
];

interface LogEntriesResult {
  LogEntries: LogEntry[],
  TotalCount: number,
}

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

async function getLogs(pageIndex: number): Promise<LogEntriesResult> {
  const resp = await fetch("http://localhost:7160/api/logs?DBName=LOG_DL&TableName=2022&PageSize=10&PageIndex=" + pageIndex);
  const jsonStr = await resp.text();
  const r = JSON.parse(jsonStr);

  // console.log(r);

  return r;
}

export default function HomePage() {
  const [pageIndex, setPageIndex] = useState<number>(1);

  const rs = useRequest(async () => {
    return await getLogs(pageIndex);
  });
  if (rs.loading) {
    return <Spin size="large" />;
  }

  const { TotalCount, LogEntries } = rs.data!;

  return <Table rowKey="ID"
    size="small"
    dataSource={LogEntries}
    columns={columns}
    pagination={{
      total: TotalCount,
      current: pageIndex,
      onChange: i => setPageIndex(i),
    }} />;
}
