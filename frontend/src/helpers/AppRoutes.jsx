import React from 'react';
import {
    Route,
    Switch,
    Redirect,
} from 'react-router-dom';
import Profile from '../pages/profile/Profile';
import Markets from '../pages/markets/Markets';
import Polls from '../pages/polls/Polls';
import Notifications from '../pages/notifications/Notifications';
import Create from '../pages/create/Create';
import About from '../pages/about/About';
import MarketDetails from '../pages/marketDetails/MarketDetails';
import User from '../pages/user/User';
import Style from '../pages/style/Style';

const AppRoutes = ({ isLoggedIn }) => {
    return (
        <Switch>
        <Route exact path='/'>
            {isLoggedIn ? <Redirect to='/markets' /> : <Redirect to='/login' />}
        </Route>
        <Route path='/profile'>
            {isLoggedIn ? <Profile /> : <Redirect to='/login' />}
        </Route>
        <Route path='/markets/:marketId' component={MarketDetails} />
        <Route path='/markets'>
            {isLoggedIn ? <Markets /> : <Redirect to='/login' />}
        </Route>
        <Route path='/polls'>
            {isLoggedIn ? <Polls /> : <Redirect to='/login' />}
        </Route>
        <Route path='/user/:username' component={User} />
        <Route path='/notifications'>
            {isLoggedIn ? <Notifications /> : <Redirect to='/login' />}
        </Route>
        <Route path='/create'>
            {isLoggedIn ? <Create /> : <Redirect to='/login' />}
        </Route>
        <Route path='/about' component={About} />
        <Route path='/style' component={Style} />
        {/* Define other routes as needed */}
        {/* ... */}
        {/* If no other route matches, redirect to home or login */}
        <Route render={() => isLoggedIn ? <Redirect to='/' /> : <Redirect to='/login' />} />
        </Switch>
    );
};

export default AppRoutes;