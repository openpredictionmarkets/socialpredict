import React from 'react';
import BrierFoxLogoPNG from '../../assets/png/BrierFoxLogo.png';

function Home() {
    return (
        <div className="min-h-screen bg-primary-background text-custom-gray-verylight">
        <div className="container mx-auto p-8">
                <header className="text-center my-12">
                    <h1 className="text-4xl font-bold text-custom-gray-light">BrierFoxForecast (BFF)</h1>
                    <p className="text-lg mt-4">An alpha project powered by SocialPredict’s open-source prediction market platform.</p>
                </header>
            <table className="w-full border-collapse">
                <tbody>
                    <tr className="border-0">
                        <td className="align-top p-4">
                            <img
                            src={BrierFoxLogoPNG}
                            alt="BrierFoxForecast Logo"
                            />
                        </td>
                        <td className="pl-0 py-4">
                            <section className="mb-12">
                            <h2 className="text-2xl font-semibold mb-4 text-custom-gray-light">About BFF</h2>
                            <p>BFF is a platform to ask questions around politics, finance, business, world news, and more. Prediction markets have proven to be one of the most accurate forms of predicting events such as USA presidential elections and geopolitical events. We are in development stages, and your questions will help build out a prediction platform that works for you.</p>
                            </section>
                            <section className="mb-12">
                            <h2 className="text-2xl font-semibold mb-4 text-custom-gray-light">Code of Conduct</h2>
                            <p>SocialPredict values free speech and freedom of expression; however, instances of blatant harassing, bullying, and racism will not be tolerated. Please keep the following rules in mind:</p>
                            <ul className="list-disc list-inside">
                                <li>No blatant racism</li>
                                <li>No advertising/solicitation</li>
                                <li>No spamming</li>
                                <li>No harassing or bullying</li>
                            </ul>
                            </section>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
            <section className="text-center">
                        <h2 className="text-2xl font-semibold mb-4 text-custom-gray-light">Join the Alpha Test</h2>
                        <p>To join the alpha test, please submit your email and desired username.</p>
                        <a href="https://docs.google.com/forms/d/1YHPWXWFpVqIvFQHz-eGPQ8f4CMuFeQ4YUWa2jS5apKw/viewform?edit_requested=true" className="text-blue-500 hover:underline">Submit here</a>
            </section>
        </div>
    );
}

export default Home;
