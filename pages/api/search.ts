import { NextApiRequest, NextApiResponse } from 'next'

const apiEndpoint = "https://operand.ai/api/search"

export default async (req: NextApiRequest, res: NextApiResponse) => {
  const {q: query} = req.query;
  const response = await fetch(apiEndpoint, {
    method: 'POST',
    headers: {
      'X-Operand-API-Key': process.env.OPERAND_API_KEY as string,
      'X-Operand-Project-ID': process.env.OPERAND_PROJECT_ID as string,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      query: query as string,
      documents: 5,
      samples: 1,
    }),
  });
  const data = await response.json();
  res.status(200).json(data);
}
