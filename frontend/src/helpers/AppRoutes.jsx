import React, { useEffect } from 'react';
import { Route, Switch, Redirect } from 'react-router-dom';
import { useAuth } from './AuthContent';
import ChangePassword from '../pages/changepassword/ChangePassword';
import Profile from '../pages/profile/Profile';
import Markets from '../pages/markets/Markets';
import Polls from '../pages/polls/Polls';
import Notifications from '../pages/notifications/Notifications';
import Create from '../pages/create/Create';
import About from '../pages/about/About';
import Stats from '../pages/stats/Stats';
import Home from '../pages/home/Home';
import MarketDetails from '../pages/marketDetails/MarketDetails';
import User from '../pages/user/User';
import Style from '../pages/style/Style';
import AdminDashboard from '../pages/admin/AdminDashboard';
import NotFound from '../pages/notfound/NotFound';

const AppRoutes = () => {
  const auth = useAuth();

  const isLoggedIn = !!auth.username;
  const isRegularUser = isLoggedIn && auth.usertype !== 'ADMIN';
  const mustChangePassword = isLoggedIn && auth.changePasswordNeeded;

  return (
    <Switch>
      {/* Stylepage */}
      <Route exact path='/style' component={Style} />

      {/* Public Routes */}
      <Route exact path='/about'>
        {isLoggedIn && mustChangePassword ? (
          <Redirect to='/changepassword' />
        ) : (
          <About />
        )}
      </Route>
      <Route exact path='/markets/:marketId'>
        {isLoggedIn && mustChangePassword ? (
          <Redirect to='/changepassword' />
        ) : (
          <MarketDetails />
        )}
      </Route>
      <Route exact path='/markets'>
        {isLoggedIn && mustChangePassword ? (
          <Redirect to='/changepassword' />
        ) : (
          <Markets />
        )}
      </Route>
      <Route exact path='/polls'>
        {isLoggedIn && mustChangePassword ? (
          <Redirect to='/changepassword' />
        ) : (
          <Polls />
        )}
      </Route>
      <Route exact path='/user/:username'>
        {isLoggedIn && mustChangePassword ? (
          <Redirect to='/changepassword' />
        ) : (
          <User />
        )}
      </Route>
      <Route exact path='/stats'>
        {isLoggedIn && mustChangePassword ? (
          <Redirect to='/changepassword' />
        ) : (
          <Stats />
        )}
      </Route>

      {/* Private Routes for Regular Users Only */}
      <Route exact path='/changepassword'>
	{isLoggedIn ? <ChangePassword /> : <Redirect to='/' />}
      </Route>
      <Route exact path='/create'>
        {!isLoggedIn ? (
          <Redirect to='/' />
        ) : mustChangePassword ? (
          <Redirect to='/changepassword' />
        ) : isRegularUser ? (
          <Create />
        ) : (
          <Redirect to='/' />
        )}
      </Route>
      <Route exact path='/notifications'>
        {!isLoggedIn ? (
          <Redirect to='/' />
        ) : mustChangePassword ? (
          <Redirect to='/changepassword' />
        ) : isRegularUser ? (
          <Notifications />
        ) : (
          <Redirect to='/' />
        )}
      </Route>
      <Route exact path='/profile'>
        {isRegularUser ? <Profile /> : <Redirect to='/' />}
      </Route>

      {/* Admin Routes */}
      <Route exact path='/admin'>
        {isLoggedIn && mustChangePassword ? (
	  <Redirect to='/changepassword' />
	) : isLoggedIn && auth.usertype === 'ADMIN' ? (
          <AdminDashboard />
        ) : (
          <Redirect to='/' />
        )}
      </Route>

      {/* Home Route */}
      <Route exact path='/'>
        {isLoggedIn && mustChangePassword ? (
          <Redirect to='/changepassword' />
        ) : (
          <Home />
        )}
      </Route>

      {/* 404 Route - This should be the last route */}
      <Route path='*'>
        <NotFound />
      </Route>
    </Switch>
  );
};

export default AppRoutes;
