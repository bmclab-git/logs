import { Table } from "antd";
import { useEffect } from "react";

const dataSource = [
  {
    key: '1',
    name: '胡彦斌',
    age: 32,
    address: '西湖区湖底公园1号',
  },
  {
    key: '2',
    name: '胡彦祖',
    age: 42,
    address: '西湖区湖底公园1号',
  },
];

const columns = [
  {
    title: '姓名',
    dataIndex: 'name',
    key: 'name',
  },
  {
    title: '年龄',
    dataIndex: 'age',
    key: 'age',
  },
  {
    title: '住址',
    dataIndex: 'address',
    key: 'address',
  },
];


export default function HomePage() {
  useEffect(() => {
    // // Copy
    // document.oncopy = () => dispatch({ type: 'keyListVM/copy', });
    // // Paste
    // document.onpaste = e => {
    //     const clipboardData = e.clipboardData;
    //     if (!clipboardData) return;

    //     const clipboardText = clipboardData?.getData("text");
    //     if (!clipboardText || clipboardText.indexOf(u.CLIPBOARD_REDIS) !== 0) {
    //         return;
    //     }

    //     Modal.confirm({
    //         title: 'Caution!',
    //         content: 'If key exists, it will be overrided, continue?',
    //         onOk() {
    //             dispatch({
    //                 type: 'keyListVM/paste',
    //                 clipboardText: clipboardText,
    //             });
    //         },
    //     });
    // };
  }, []);

  return (
    <div>
      <Table size="small" dataSource={dataSource} columns={columns} />
    </div>
  )
}
