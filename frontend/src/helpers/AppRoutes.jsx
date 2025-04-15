import { Route, Routes, Navigate } from 'react-router-dom';
import { useAuth } from './AuthContent.jsx';
import ChangePassword from '../pages/changepassword/ChangePassword.jsx';
import Profile from '../pages/profile/Profile.jsx';
import Markets from '../pages/markets/Markets.jsx';
import Polls from '../pages/polls/Polls.jsx';
import Notifications from '../pages/notifications/Notifications.jsx';
import Create from '../pages/create/Create.jsx';
import About from '../pages/about/About.jsx';
import Home from '../pages/home/Home.jsx';
import MarketDetails from '../pages/marketDetails/MarketDetails.jsx';
import User from '../pages/user/User.jsx';
import Style from '../pages/style/Style.jsx';
import AdminDashboard from '../pages/admin/AdminDashboard.jsx';
import NotFound from '../pages/notfound/NotFound.jsx';

const AppRoutes = () => {
  const auth = useAuth();

  const isLoggedIn = !!auth.username;
  const isRegularUser = isLoggedIn && auth.usertype !== 'ADMIN';
  const mustChangePassword = isLoggedIn && auth.changePasswordNeeded;

  return (
    <Routes>
      {/* Stylepage */}
      <Route exact path='/style' component={Style} />

      {/* Public Routes */}
      <Route exact path='/about'>
        {isLoggedIn && auth.usertype !== 'ADMIN' && mustChangePassword ? (
          <Navigate to='/changepassword' />
        ) : (
          <About />
        )}
      </Route>
      <Route exact path='/markets/:marketId'>
        {isLoggedIn && auth.usertype !== 'ADMIN' && mustChangePassword ? (
          <Navigate to='/changepassword' />
        ) : (
          <MarketDetails />
        )}
      </Route>
      <Route exact path='/markets'>
        {isLoggedIn && auth.usertype !== 'ADMIN' && mustChangePassword ? (
          <Navigate to='/changepassword' />
        ) : (
          <Markets />
        )}
      </Route>
      <Route exact path='/polls'>
        {isLoggedIn && auth.usertype !== 'ADMIN' && mustChangePassword ? (
          <Navigate to='/changepassword' />
        ) : (
          <Polls />
        )}
      </Route>
      <Route exact path='/user/:username'>
        {isLoggedIn && auth.usertype !== 'ADMIN' && mustChangePassword ? (
          <Navigate to='/changepassword' />
        ) : (
          <User />
        )}
      </Route>

      {/* Private Routes for Regular Users Only */}
      <Route exact path='/changepassword'>
        {isRegularUser ? <ChangePassword /> : <Navigate to='/' />}
      </Route>
      <Route exact path='/create'>
        {!isLoggedIn ? (
          <Navigate to='/' />
        ) : mustChangePassword ? (
          <Navigate to='/changepassword' />
        ) : isRegularUser ? (
          <Create />
        ) : (
          <RediNavigaterect to='/' />
        )}
      </Route>
      <Route exact path='/notifications'>
        {!isLoggedIn ? (
          <Navigate to='/' />
        ) : mustChangePassword ? (
          <Navigate to='/changepassword' />
        ) : isRegularUser ? (
          <Notifications />
        ) : (
          <Navigate to='/' />
        )}
      </Route>
      <Route exact path='/profile'>
        {isRegularUser ? <Profile /> : <Navigate to='/' />}
      </Route>

      {/* Admin Routes */}
      <Route exact path='/admin'>
        {isLoggedIn && auth.usertype === 'ADMIN' ? (
          <AdminDashboard />
        ) : (
          <Navigate to='/' />
        )}
      </Route>

      {/* Home Route */}
      <Route exact path='/'>
        {isLoggedIn && auth.usertype !== 'ADMIN' && mustChangePassword ? (
          <Navigate to='/changepassword' />
        ) : (
          <Home />
        )}
      </Route>

      {/* 404 Route - This should be the last route */}
      <Route path='*'>
        <NotFound />
      </Route>
    </Routes>
  );
};

export default AppRoutes;
