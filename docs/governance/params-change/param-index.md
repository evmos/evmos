---
order: 2
---

# Index of Governable Parameters

## Querying on-chain parameters

Given a subspace and an associated key, you can query on chain parameters using the CLI.

``` bash
evmosd query params subspace <subspace_name> <key> --node <node_address> --chain-id <chain_id>
```

For more information on specific modules, refer to the [Cosmos SDK documentation on modules](https://docs.cosmos.network/master/).

## Current subspaces, keys, and values

<section v-for="(value, subspace) in $themeConfig.currentParameters">
   <h2><code>{{subspace}}</code> subspace</h2>
   <table>
      <tr>
         <th>Key</th>
         <th>Value</th>
      </tr>
      <tr v-for="(v,k) in value">
         <td><code>{{ k }}</code></td>
         <td><code>{{ v }}</code></td>
      </tr>
   </table>
   <p>
     Read more about the governance implications of the  <a :href="subspace + '.html'">{{subspace}} subspace here.</a>
   </p>
</section>
