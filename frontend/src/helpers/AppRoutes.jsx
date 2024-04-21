import React from 'react';
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

const AppRoutes = () => {

    const { isLoggedIn } = useAuth();

    return (
        <Switch>
        {/* Public Routes */}
        <Route exact path='/' />
        <Route path='/about' component={About} />
        <Route path='/markets/:marketId' component={MarketDetails} />
        <Route path='/markets' component={Markets}/>
        <Route path='/polls' component={Polls}/>
        <Route path='/user/:username' component={User} />
        <Route path='/style' component={Style} />
        {/* Private Routes */}
        <Route path='/create'>
            {isLoggedIn ? <Create /> : <Redirect to='/' />}
        </Route>
        <Route path='/notifications'>
            {isLoggedIn ? <Notifications /> : <Redirect to='/' />}
        </Route>
        <Route path='/profile'>
            {isLoggedIn ? <Profile /> : <Redirect to='/' />}
        </Route>
        {/* If no other route matches, redirect to home or login */}
        <Route render={() => isLoggedIn ? <Redirect to='/' /> : <Redirect to='/' />} />
        </Switch>
    );
};

export default AppRoutes;