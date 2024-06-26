import React, { useEffect } from 'react';
import {
    Route,
    Switch,
    Redirect,
} from 'react-router-dom';
import { useAuth } from './AuthContent';
import ChangePassword from '../pages/changepassword/ChangePassword';
import Profile from '../pages/profile/Profile';
import Markets from '../pages/markets/Markets';
import Polls from '../pages/polls/Polls';
import Notifications from '../pages/notifications/Notifications';
import Create from '../pages/create/Create';
import About from '../pages/about/About';
import Home from '../pages/home/Home'
import MarketDetails from '../pages/marketDetails/MarketDetails';
import User from '../pages/user/User';
import Style from '../pages/style/Style';
import AdminDashboard from '../pages/admin/AdminDashboard';

const AppRoutes = () => {

    const auth = useAuth();

    const isLoggedIn = !!auth.username;
    const isRegularUser = isLoggedIn && auth.usertype !== 'ADMIN';
    const mustChangePassword = isLoggedIn && auth.changePasswordNeeded;

    return (
        <Switch>
            {/* Stylepage */}
            <Route path='/style' component={Style} />

            {/* Public Routes */}
            <Route path='/about'>
                {isLoggedIn && !auth.usertype === 'ADMIN' && mustChangePassword ? (
                    <Redirect to='/changepassword' />
                ) : (
                    <About />
                )}
            </Route>
            <Route path='/markets/:marketId'>
            {isLoggedIn && !auth.usertype === 'ADMIN' && mustChangePassword ? (
                    <Redirect to='/changepassword' />
                ) : (
                    <MarketDetails />
                )}
            </Route>
            <Route path='/markets'>
            {isLoggedIn && !auth.usertype === 'ADMIN' && mustChangePassword ? (
                    <Redirect to='/changepassword' />
                ) : (
                    <Markets />
                )}
            </Route>
            <Route path='/polls'>
            {isLoggedIn && !auth.usertype === 'ADMIN' && mustChangePassword ? (
                    <Redirect to='/changepassword' />
                ) : (
                    <Polls />
                )}
            </Route>
            <Route path='/user/:username'>
            {isLoggedIn && !auth.usertype === 'ADMIN' && mustChangePassword ? (
                    <Redirect to='/changepassword' />
                ) : (
                    <User />
                )}
            </Route>

            {/* Private Routes for Regular Users Only */}
            <Route path='/changepassword'>
                {isRegularUser ? <ChangePassword /> : <Home />}
            </Route>
            <Route path='/create'>
                {!isLoggedIn ? (
                        <Home />
                    ) : mustChangePassword ? (
                        <Redirect to='/changepassword' />
                    ) : isRegularUser ? (
                        <Create />
                    ) : (
                        <Home /> // catch all for all other condtions, not a regular user
                    )}
            </Route>
            <Route path='/notifications'>
                {!isLoggedIn ? (
                    <Home />
                ) : mustChangePassword ? (
                    <Redirect to='/changepassword' />
                ) : isRegularUser ? (
                    <Notifications />
                ) : (
                    <Home /> // catch all for all other condtions, not a regular user
                )}
            </Route>
            <Route path='/profile'>
                {isRegularUser ? <Profile /> : <Home />}
            </Route>
            {/* Admin Routes */}
            <Route path='/admin'>
                {isLoggedIn && auth.usertype === 'ADMIN' ? <AdminDashboard /> : <Home />}
            </Route>

            {/* If no other route matches, redirect to home */}
            <Route path='/'>
            {isLoggedIn && !auth.usertype === 'ADMIN' && mustChangePassword ? (
                    <Redirect to='/changepassword' />
                ) : (
                    <Home />
                )}
            </Route>

        </Switch>
    );
};

export default AppRoutes;
