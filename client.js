const axios = require("axios");
const fs = require('fs');
const path = require('path');
const readline = require('readline');

const questionBuilder = async () => {
    let questions = [];
    try {
        if(fs.existsSync(path.join(__dirname, './questions.txt'))) {
            const fileStream = fs.createReadStream(path.join(__dirname, './questions.txt'), 'utf-8');

            const rl = readline.createInterface({
                input: fileStream,
                crlfDelay: Infinity
            });

            for await (const line of rl) {
                const lineParts = line.split(',');
                questions.push(lineParts[0].trim());
            }
        }
        return questions;

    } catch (err) {
        console.error("Error reading questions file:", err);
    }
    return questions;
}

// const outputBuilder = async (questions) => {
//     questions.forEach(async question => {
//         console.log(question);
//         console.log(`Processing question: ${question}`);
//         await axios.get(`http://localhost:3000/campaigns/vhf4wt/${encodeURIComponent(question)}`)
//             .then(response => {
//                 console.log(`Response for question "${question}":`, response.data);
//             })
//             .catch(error => {
//                 console.error(`Error fetching data for question "${question}":`, error.message, error.response.data);
//             });
//     });
// }

const outputBuilder = async (questions) => {
    for (const question of questions) {
        console.log(`Processing question: ${question}`);
        
        try {
            const response = await axios.get(`http://localhost:3000/campaigns/bssqmz`, {
                params: { text: question }
            });
            console.log(`Success for "${question}":`, response.data);
        } catch (error) {
            console.error(`Error for "${question}":`, 
                error.response ? error.response.data : error.message
            );
        }
    }
}

(async () => {
    const questions = await questionBuilder();
    console.log(`Total questions to process: ${questions.length}`);
    await outputBuilder(questions);
    console.log("All questions processed.");
})();
