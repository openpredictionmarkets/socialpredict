import React from 'react';

function About() {
  return (
    <div className='flex flex-col h-full text-white p-4 md:p-6 lg:p-8'>
      <h1 className='text-2xl md:text-3xl font-bold mb-4'>
        About SocialPredict
      </h1>
      <div className='bg-gray-800 rounded-lg shadow-lg flex-grow overflow-auto'>
        <div className='p-4 md:p-6 space-y-4 md:space-y-6'>
          <p className='text-center'>
            <a
              href='https://github.com/openpredictionmarkets/socialpredict'
              target='_blank'
              rel='noopener noreferrer'
              className='text-blue-400 hover:text-blue-300 transition-colors duration-200'
            >
              📈 Built with SocialPredict <br /> ⭐ Star Us on Github!
            </a>
          </p>

          <section>
            <h2 className='text-xl font-semibold mb-2'>
              Empowering Communities with Domain-Specific Insights
            </h2>
            <p className='mb-2'>Have you heard of Prediction Markets?</p>
            <ul className='list-disc list-inside pl-4'>
              <li>
                Prediction Markets, unlike polls, incentivize accuracy.
                Participants take a stake in what they think is correct,
                promoting rigorous research and reducing bias.
              </li>
            </ul>
          </section>

          <section>
            <h3 className='text-lg font-semibold mb-2'>
              Efficiency through Community Engagement
            </h3>
            <p className='mb-2'>SocialPredict is Open Source Software Which:</p>
            <ul className='list-disc list-inside pl-4'>
              <li>
                Embraces the open-source ethos, making our platform free for
                anyone to deploy under the MIT License.
              </li>
            </ul>
          </section>

          <section>
            <h3 className='text-lg font-semibold mb-2'>
              Domain-Specific Prediction Markets
            </h3>
            <p className='mb-2'>
              Imagine a prediction market platform tailored to specific
              interests, for example, photography and cameras:
            </p>
            <ul className='list-disc list-inside pl-4'>
              <li>
                An admin runs a photographers-and-industry-specialists-only
                prediction market platform.
              </li>
              <li>
                Discussions and bets on technology predictions, specific to
                photography and adjacent technology.
              </li>
            </ul>
          </section>

          <section>
            <h3 className='text-lg font-semibold mb-2'>
              Community Empowering Mission
            </h3>
            <p className='mb-2'>We strive to:</p>
            <ul className='list-disc list-inside pl-4'>
              <li>Empower communities to predict outcomes efficiently.</li>
              <li>Foster deeper understanding of chosen domains.</li>
              <li>Facilitate the exchange of valuable insights.</li>
            </ul>
          </section>

          <p className='text-lg font-bold text-center my-4 md:my-6'>
            Join us in shaping the future of prediction markets by building
            connections and expertise within your community.
          </p>

          <hr className='border-gray-700 my-4 md:my-6' />

          <section>
            <h2 className='text-xl font-bold mb-2'>Join Us</h2>
            <p className='mb-2'>There are a few ways to support us:</p>
            <ul className='list-disc list-inside pl-4 space-y-2'>
              <li>
                <a
                  href='https://github.com/openpredictionmarkets/socialpredict'
                  target='_blank'
                  rel='noopener noreferrer'
                  className='text-blue-400 hover:text-blue-300 transition-colors duration-200'
                >
                  ⭐ Star Us on Github!
                </a>
              </li>
              <li>
                <a
                  href='https://github.com/openpredictionmarkets/socialpredict/issues'
                  target='_blank'
                  rel='noopener noreferrer'
                  className='text-blue-400 hover:text-blue-300 transition-colors duration-200'
                >
                  🔧 Submit an Issue on Github!
                </a>
              </li>
              <li>
                <a
                  href='https://github.com/openpredictionmarkets/socialpredict/blob/main/README.md'
                  target='_blank'
                  rel='noopener noreferrer'
                  className='text-blue-400 hover:text-blue-300 transition-colors duration-200'
                >
                  ☁️ Spin Up Your Own Instance Following Our Instructions!
                </a>
              </li>
            </ul>
          </section>
        </div>
      </div>
    </div>
  );
}

export default About;
