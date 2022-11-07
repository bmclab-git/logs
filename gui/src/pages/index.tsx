import { Spin, Table } from "antd";
import { useRequest } from 'ahooks';
import { useState } from "react";

const columns = [
  {
    title: 'Level',
    dataIndex: 'Level',
  },
  {
    title: 'Message',
    dataIndex: 'Message',
  },
  {
    title: 'User',
    dataIndex: 'User',
  },
  {
    title: 'TraceNo',
    dataIndex: 'TraceNo',
  },
  {
    title: 'Date',
    dataIndex: 'CreatedOnUtc',
  },
];

interface LogEntriesResult {
  LogEntires: LogEntry[],
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
  const resp = await fetch("http://localhost:8080/api/logs?DBName=LOG_DL&TableName=2022&PageSize=10&PageIndex=" + pageIndex);
  const jsonStr = await resp.text();
  const r = JSON.parse(jsonStr);

  return r;
}

export default function HomePage() {
  const [pageIndex, setPageIndex] = useState<number>(1);

  const rs = useRequest(async () => {
    await getLogs(pageIndex);
  });
  if (rs.loading) {
    return <Spin size="large" />;
  }

  const { TotalCount, LogEntires } = rs.data!;

  return <Table size="small" dataSource={LogEntires} columns={columns} pagination={{
    total: 100,
    current: pageIndex,
    onChange: i => setPageIndex(i),
  }} />;
}
