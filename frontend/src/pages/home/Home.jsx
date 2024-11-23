import React from 'react';

function Home() {
  return (
    <div className='min-h-[calc(100vh-6rem)] bg-primary-background text-custom-gray-verylight flex flex-col justify-center py-8 px-4 sm:px-6 lg:px-8'>
      <div className='max-w-4xl mx-auto w-full'>
        <div className='flex flex-col sm:flex-row items-center mb-8'>
          <img
            src='../../assets/png/BrierFoxLogo.png'
            alt='BrierFoxForecast Logo'
            className='w-24 h-24 sm:w-32 sm:h-32 mb-4 sm:mb-0 sm:mr-6'
          />
          <div className='flex flex-col justify-center h-full text-center sm:text-left'>
            <h1 className='text-3xl sm:text-4xl font-bold text-custom-gray-light mb-2'>
              BrierFoxForecast (BFF)
            </h1>
            <p className='text-lg text-custom-gray-light'>
              An alpha project powered by SocialPredict's open-source prediction
              market platform.
            </p>
          </div>
        </div>

        <div className='space-y-8'>
          <section className='bg-gray-800 rounded-lg p-6 shadow-lg'>
            <h2 className='text-xl font-semibold mb-3 text-custom-gray-light'>
              About BFF
            </h2>
            <p className='text-base mb-4'>
              BFF is a platform for predictions on politics, finance, business,
              world news, and more. We're in development, and your input will
              shape a prediction platform that works for you.
            </p>
            <h3 className='text-lg font-semibold mb-2 text-custom-gray-light'>
              Code of Conduct
            </h3>
            <p className='text-base mb-2'>
              We value free speech but do not tolerate:
            </p>
            <ul className='list-disc list-inside text-base pl-4'>
              <li>Blatant racism</li>
              <li>Advertising or solicitation</li>
              <li>Spamming</li>
              <li>Harassment or bullying</li>
            </ul>
          </section>

          <section className='text-center bg-blue-600 p-6 rounded-lg shadow-lg'>
            <h2 className='text-xl font-semibold mb-3 text-white'>
              Join the Alpha Test
            </h2>
            <p className='text-base mb-4 text-white'>
              Submit your email and desired username to participate.
            </p>
            <a
              href='https://docs.google.com/forms/d/1YHPWXWFpVqIvFQHz-eGPQ8f4CMuFeQ4YUWa2jS5apKw/viewform?edit_requested=true'
              className='inline-block bg-white text-blue-600 py-2 px-4 rounded-lg font-semibold text-base hover:bg-gray-100 transition duration-300'
            >
              Submit Your Application
            </a>
          </section>
        </div>
      </div>
    </div>
  );
}

export default Home;
