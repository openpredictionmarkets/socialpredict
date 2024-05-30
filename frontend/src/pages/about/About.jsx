import React from 'react';

function About() {
    return (
        <div className="p-6 bg-primary-background shadow-md rounded-lg text-white">
            <h1 className="text-2xl font-bold mb-4">About SocialPredict</h1>
            <p className="text-sm text-center p-4"><a href="https://github.com/openpredictionmarkets/socialpredict" target="_blank" rel="noopener noreferrer" className="hover:text-blue-700">📈 Built with SocialPredict ⭐ Star Us on Github!</a></p>
            <div className="space-y-4">
                <h2 className="text-xl font-semibold">Empowering Communities with Domain-Specific Insights</h2>
                <p>Have you heard of Prediction Markets?</p>
                <ul className="list-disc list-inside">
                    <li>Prediction Markets, unlike polls, incentivize accuracy. Participants take a stake in what they think is correct, promoting rigorous research and reducing bias.</li>
                </ul>

                <h3 className="text-lg font-semibold">Efficiency through Community Engagement</h3>
                <p>SocialPredict is Open Source Software Which:</p>
                <ul className="list-disc list-inside">
                    <li>Embraces the open-source ethos, making our platform free for anyone to deploy under the MIT License.</li>
                </ul>

                <h3 className="text-lg font-semibold">Domain-Specific Prediction Markets</h3>
                <p>Imagine a prediction market platform tailored to specific interests, for example, photography and cameras:</p>
                <ul className="list-disc list-inside">
                    <li>An admin runs a photographers-and-industry-specialists-only prediction market platform.</li>
                    <li>Discussions and bets on technology predictions, specific to photography and adjacent technology.</li>
                </ul>

                <h3 className="text-lg font-semibold">Comunity Empowering Mission</h3>
                We strive to:
                <ul className="list-disc list-inside">
                    <li>Empower communities to predict outcomes efficiently.</li>
                    <li>Foster deeper understanding of chosen domains.</li>
                    <li>Facilitate the exchange of valuable insights.</li>
                </ul>

                <h2 className="text-xl font-bold">Join us in shaping the future of prediction markets by building connections and expertise within your community.</h2>
                <hr></hr>
                <h2 className="text-xl font-bold">Join Us</h2>
                There are a few ways to support us.
                <ul className="list-disc list-inside">
                    <li><a href="https://github.com/openpredictionmarkets/socialpredict" target="_blank" rel="noopener noreferrer" className="hover:text-blue-700">⭐ Star Us on Github!</a></li>
                    <li><a href="https://github.com/openpredictionmarkets/socialpredict/issues" target="_blank" rel="noopener noreferrer" className="hover:text-blue-700">🔧 Submit an Issue on Github!</a></li>
                    <li><a href="https://github.com/openpredictionmarkets/socialpredict/blob/main/README.md" target="_blank" rel="noopener noreferrer" className="hover:text-blue-700">☁️ Spin Up Your Own Instance Following Our Instructions!</a></li>
                </ul>
            </div>
        </div>
    );
}

export default About;
