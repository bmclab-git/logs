import { Outlet } from 'umi';
import 'antd/dist/antd.css';

export default function Layout() {
  return <div id="container">
    <Outlet />
    </div>
}
