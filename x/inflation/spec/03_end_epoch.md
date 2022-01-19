<!--
order: 3
-->

# Begin-Epoch

Minting parameters are recalculated and inflation
paid at the beginning of each epoch. An epoch is signalled by x/epochs

## NextEpochProvisions

The target epoch provision is recalculated on each reduction period (default 3 years).
At the time of reduction, the current provision is multiplied by reduction factor (default `2/3`),
to calculate the provisions for the next epoch. Consequently, the rewards of the next period
will be lowered by `1 - reduction factor`.

```go
func (m Minter) NextEpochProvisions(params Params) sdk.Dec {
    return m.EpochProvisions.Mul(params.ReductionFactor)
}
```

## EpochProvision

Calculate the provisions generated for each epoch based on current epoch provisions. The provisions are then minted by the `mint` module's `ModuleMinterAccount`. These rewards are transferred to a `FeeCollector`, which handles distributing the rewards per the chains needs. (See TODO.md for details) This fee collector is specified as the `auth` module's `FeeCollector` `ModuleAccount`.

```go
func (m Minter) EpochProvision(params Params) sdk.Coin {
    provisionAmt := m.EpochProvisions.QuoInt(sdk.NewInt(int64(params.EpochsPerYear)))
    return sdk.NewCoin(params.MintDenom, provisionAmt.TruncateInt())
}
```
