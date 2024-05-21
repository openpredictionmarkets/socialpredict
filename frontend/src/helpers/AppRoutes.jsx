import React, { useEffect } from 'react';
import {
    Route,
    Switch,
    Redirect,
} from 'react-router-dom';
import { useAuth } from './AuthContent';
import Profile from '../pages/profile/Profile';
import Markets from '../pages/markets/Markets';
import Polls from '../pages/polls/Polls';
import Notifications from '../pages/notifications/Notifications';
import Create from '../pages/create/Create';
import About from '../pages/about/About';
import MarketDetails from '../pages/marketDetails/MarketDetails';
import User from '../pages/user/User';
import Style from '../pages/style/Style';
import AdminDashboard from '../pages/admin/AdminDashboard';

const AppRoutes = () => {

    const auth = useAuth();

    console.log("Auth state: ", auth);

    const isLoggedIn = !!auth.username;
    const isRegularUser = isLoggedIn && auth.usertype !== 'ADMIN';

    console.log("user type: ", auth.usertype);

    return (
        <Switch>
            {/* Public Routes */}
            <Route path='/about' component={About} />
            <Route path='/markets/:marketId' component={MarketDetails} />
            <Route path='/markets' component={Markets} />
            <Route path='/polls' component={Polls} />
            <Route path='/user/:username' component={User} />
            <Route path='/style' component={Style} />
            {/* Private Routes for Regular Users Only */}
            <Route path='/create'>
                {isRegularUser ? <Create /> : <Redirect to='/' />}
            </Route>
            <Route path='/notifications'>
                {isRegularUser ? <Notifications /> : <Redirect to='/' />}
            </Route>
            <Route path='/profile'>
                {isRegularUser ? <Profile /> : <Redirect to='/' />}
            </Route>
            {/* Admin Routes */}
            <Route path='/admin'>
                {isLoggedIn && auth.usertype === 'ADMIN' ? <AdminDashboard /> : <Redirect to='/' />}
            </Route>
            {/* If no other route matches, redirect to home */}
            <Route render={() => <Redirect to='/' />} />
        </Switch>
    );
};

export default AppRoutes;
