let numberOfYears: number = 50;
let a_initialValue: number = 1000000000;
let r_decayFactor: number = 0.2745447257;
let c_longTermSupply: number = 0;   //constant inflation
//fraction of the staking tokens which are currently bonded
let bondedRatio: number = 0.66;
//the max amount to increase inflation
let maxVarience: number = 0;
//our optimal bonded ratio
let bondingTarget: number = 0.66;

let x_startingYear = 5;
console.log("Initial supply: " + a_initialValue);

for(let i: number = x_startingYear; i < numberOfYears + x_startingYear; i++) {
    // exponentialDecay calculations
    let yearlyReduceCoeff: number = 1 - r_decayFactor;
    yearlyReduceCoeff = yearlyReduceCoeff ** i;
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
    let growPart: number = yearInflationAmount / a_initialValue;
    a_initialValue = a_initialValue + yearInflationAmount;
    console.log("year " + (i+1-x_startingYear) + "; TotalFunds " + a_initialValue + "; Yearly inflation " + yearInflationAmount + "; Supply grow part " + growPart + "; Epoch provision " + yearInflationAmount / 365);
}
// Note: please restart the page if syntax highlighting works bad.
let el = document.querySelector('#header')

let msg: string = 'Hi friend, try edit me!'
el.innerHTML = msg

console.log('it shows results as you type')