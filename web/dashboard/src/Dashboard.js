import React, { useState } from 'react';
import OpenAPIClientAxios from 'openapi-client-axios';


const api = new OpenAPIClientAxios({ definition: 'http://127.0.0.1:20000/openapi/litterboxsvc/openapi.yaml' });


function Dashboard() {
  const [data, setData] = React.useState(null)
  function click() {
    api.init()
    .then(client => client.Accounts_GetAccount('default')).then(res => console.log(res.data) || setData(res.data))
    .catch(error => console.log(error) || setData(error))
  }

  return (
    <div className="Dashboard space-y-8">
      <button onClick={click}
        type="button"
        className="inline-flex items-center px-2.5 py-1.5 border border-transparent text-xs font-medium rounded shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
      >
        Get Default Account
      </button>
      <pre className="font-mono bg-gray-600 text-white p-4 rounded-md">
        {JSON.stringify(data, null, 2)}
      </pre>
    </div>
  );
}

export default Dashboard;
