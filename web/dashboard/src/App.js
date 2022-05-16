import React from 'react';
import OpenAPIClientAxios from 'openapi-client-axios';
import './App.css';


const api = new OpenAPIClientAxios({ definition: '/openapi/accountsvc/openapi.yaml' });


function App() {
  function click() {
    api.init().then(client => client.Accounts_GetAccount('default')).then(res => console.log(res.data))
  }

  return (
    <div className="App">
      <button onClick={click}>Get Default Account</button>
    </div>
  );
}

export default App;
