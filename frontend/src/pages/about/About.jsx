import React from 'react';

function About() {
    return (
        <div className="p-6 bg-primary-background shadow-md rounded-lg text-white">
            <h1 className="text-2xl font-bold mb-4">About SocialPredict</h1>
            <div className="space-y-4">
                <h2 className="text-xl font-semibold">Empowering Communities with Domain-Specific Insights</h2>
                <p>Have you heard of Prediction Markets?</p>
                <ul className="list-disc list-inside">
                    <li>Prediction Markets, unlike polls, incentivize accuracy. Participants take a stake in what they think is correct, promoting rigorous research and reducing bias.</li>
                    <li>In contrast, social media posts and polls, using upvotes and downvotes or having users choose with no stake, are not as robust.</li>
                    <li>Of course, as with anything, the knowledge of a community is tied to the ability of the participants.</li>
                    <li>Our solution is to empower individuals to run their own prediction market platforms to attempt to use the wisdom of their own crowds to out-predict one another.</li>
                </ul>

                <h3 className="text-lg font-semibold">Efficiency through Community Engagement</h3>
                <p>SocialPredict is Open Source Software Which:</p>
                <ul className="list-disc list-inside">
                    <li>Embraces the open-source ethos, making our platform free for anyone to deploy.</li>
                    <li>Enables users to harness domain-specific expertise to create more accurate predictions.</li>
                </ul>

                <h3 className="text-lg font-semibold">Domain-Specific Prediction Markets</h3>
                <p>Imagine a prediction market platform tailored to specific interests, for example, photography and cameras:</p>
                <ul className="list-disc list-inside">
                    <li>An admin runs a photographers-and-industry-specialists-only prediction market platform.</li>
                    <li>Discussions and bets on technology predictions, specific to photography and adjacent technology, will focus on understanding what will be new in the following year.</li>
                </ul>

                <p>SocialPredict's mission is to provide a versatile platform that:</p>
                <ul className="list-disc list-inside">
                    <li>Empowers communities to predict outcomes efficiently.</li>
                    <li>Fosters a deeper understanding of chosen domains.</li>
                    <li>Facilitates the exchange of valuable insights.</li>
                </ul>

                <h2 className="text-xl font-bold">Join us in shaping the future of prediction markets by building connections and expertise within your community.</h2>
            </div>
        </div>
    );
}

export default About;
