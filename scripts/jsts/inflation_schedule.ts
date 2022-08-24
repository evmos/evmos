let numberOfYears: number = 22;
let a_initialValue: number = 1000000000;
let r_decayFactor: number = 0.558621;
let c_longTermSupply: number = 0;
//fraction of the staking tokens which are currently bonded
let bondedRatio: number = 0.66;
//the max amount to increase inflation
let maxVarience: number = 0;
//our optimal bonded ratio
let bondingTarget: number = 0.66;
console.log("Initial supply: " + a_initialValue);

for(let x: number = 1; x < numberOfYears; x++) {
    // exponentialDecay calculations
    let yearlyReduceCoeff: number = 1 - r_decayFactor;
    yearlyReduceCoeff = yearlyReduceCoeff ** x;
    //console.log("yearly reduce coeff " + yearlyReduceCoeff);

    let arx: number = a_initialValue * yearlyReduceCoeff;
    let exponentialDecay: number = arx + c_longTermSupply;
    //console.log("exponential decay " + exponentialDecay) ;

    // bondingIncentive calculation
    let mvbt: number = maxVarience / bondingTarget;
    let mvbtbr: number = bondedRatio * mvbt;
    let mv1: number = 1 + maxVarience;
    let bondingIncentive: number = mv1 - mvbtbr;
    //console.log("Bonding incentive " +bondingIncentive);
    let yearInflationAmount: number = exponentialDecay * bondingIncentive;
    a_initialValue = a_initialValue + yearInflationAmount;
    console.log("year " + (x+1) + " TotalFunds " + a_initialValue + " Yearly inflation " + yearInflationAmount + " Epoch provision " + yearInflationAmount / 365);
}