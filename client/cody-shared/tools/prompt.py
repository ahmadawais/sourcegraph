#!/usr/bin/env python3

import asyncio
import collections
import math
import os
import requests
import sys

from concurrent.futures import ThreadPoolExecutor

stock_cody_preamble = '''Human: You are Cody, an AI-powered coding assistant created by Sourcegraph. You work inside a text editor. You have access to my currently open files. You perform the following actions:
- Answer general programming questions.
- Answer questions about the code that I have provided to you.
- Generate code that matches a written description.
- Explain what a section of code does.`

In your responses, obey the following rules:
- Be as brief and concise as possible without losing clarity.
- All code snippets have to be markdown-formatted, and placed in-between triple backticks like this \`\`\`.
- Answer questions only if you know the answer or can make a well-informed guess. Otherwise, tell me you don't know and what context I need to provide you for you to answer the question.
- Only reference file names or URLs if you are sure they exist.

Assistant: Understood. I am Cody, an AI assistant made by Sourcegraph to help with programming tasks.
I work inside a text editor. I have access to your currently open files in the editor.
I will answer questions, explain code, and generate code as concisely and clearly as possible.
My responses will be formatted using Markdown syntax for code blocks.
I will acknowledge when I don't know an answer or need more context.

Human: '''

evaluation_instructions = '''The output must NOT contain <cody-help>.
The output MUST contain <cody-replace>.
There MUST be some lines of context repeated from the input after the start and before the end of the <cody-replace> tags.
'''

inputs = [
    ('',
     'document what does this program does',
     'hello.rs',
     '',
     '''
fn main() {
    println("Hello, world! From Cody demo land.")
}''',
    ''),

    ('',
     'write a doc comment explaining what this function does',
     'main.rs',
     '''    if denominator != 0 && numerator % denominator == 0 {
        numerator / denominator
    } else {
        i32::MIN
    }
}

''',
'''fn digit(n: u16) -> &'static str {
    match n {
        0b10 => ">1",
        0b100 => ">2",
        0b1000 => ">3",
        0b10000 => ">4",
        0b100000 => ">5",
        0b1000000 => ">6",
        0b10000000 => ">7",
        0b100000000 => ">8",
        0b1000000000 => ">9",
        _ if n.count_ones() == 0 => "!",
        _ if n.count_ones() == 1 => "?1",
        _ if n.count_ones() == 2 => "?2",
        _ if n.count_ones() == 3 => "?3",
        _ if n.count_ones() == 4 => "?4",
        _ if n.count_ones() == 5 => "?5",
        _ if n.count_ones() == 6 => "?6",
        _ if n.count_ones() == 7 => "?7",
        _ if n.count_ones() == 8 => "?8",
        _ if n.count_ones() == 9 => "?9",
        _ => panic!("unreachable"),
    }
}''',
'''
const DIM: usize = 9;
#[derive(Clone)]
struct Board {
    cells: [u16; DIM * DIM],
}

''')
]

# Can consult 'gen' oracle to pick the prompt it is generating.
# Returns a function which, given specific inputs, returns a tuple of whole prompt and context for the meta-evaluator.
def generate_prompt(gen):
    prompt = []
    prompt.append(gen.pick('FLATTER', ['', 'You are an expert software engineer who writes flawless code.', 'You write code for a hobby.']))
    prompt.append(gen.pick('TASK', ['', 'I need your help to improve some code.', 'This is the code you are writing.']))
    # TODO: test presence of context.responseMultiplexer.prompt()
    text = ' '.join(prompt)
    def g(input):
        context = f'<cody-replace>{input[3]}<cody-help prompt="{input[1]}">{input[4]}</cody-help>${input[5]}</cody-replace>'
        return (stock_cody_preamble + text + f'''
The area I need help with is highlighted with <cody-help> tags. You are helping me work on that part.
Follow the instructions in the prompt attribute and produce a rewritten replacement.
Strip the <cody-help> tags from your reply, just leave the improved content inside the tags.
Put the replacement in <cody-replace> tags.
I need only the replacement, no other commentary about it. Do not write anything after the closing </cody-replace> tag.
If you are adding code, I need you to repeat three lines of my code verbatim, both before and after your new code, so I understand where to insert your new code. You should put that repeated code within the <cody-replace> tags.

Assistant: OK, I understand. I will follow the prompts to improve the code, and only reply with code in <cody-replace> tags. The last thing I write will be the closing </cody-replace> tag. I will not write code outside <cody-replace> tags. I will not write the <cody-help> tags in my reply.

Human: Great, thank you! This is part of the file ${
            input[2]
        }. The area I need help with is highlighted with <cody-help> tags. Again, I only need the replacement in <cody-replace> tags.

{context}\n\nAssistant: ''', context)
    return g


class Schema(object):
    def __init__(self):
        self.dimensions = {}
        self.work = [Prompt(self, {})]
        self.prompts = []

    def discover_prompts(self, prompt_callback):
        while self.work:
            item = self.work.pop()
            callback = prompt_callback(item)
            item.id = len(self.prompts)
            item.generator = callback
            self.prompts.append(item)

    def advise(self, generator, label, n):
        if label in self.dimensions:
            assert self.dimensions[label] == n
        else:
            self.dimensions[label] = n
        if label not in generator.choices:
            for i in range(1, n):
                obj = generator.choices.copy()
                obj[label] = i
                self.work.append(Prompt(self, obj))
            generator.choices[label] = 0
        return generator.choices[label]


class Prompt(object):
    def __init__(self, schema, choices):
        self.schema = schema
        self.choices = choices or {}

    def pick(self, label, choices):
        return choices[self.schema.advise(self, label, len(choices))]

    def __repr__(self):
        return str(self.choices)


anthropic_key = os.getenv('ANTHROPIC_KEY')
if not anthropic_key:
    print('set ANTHROPIC_KEY and re-run')
    sys.exit(1)


def submit_prompt(prompt):
    """Submit a prompt to the Anthropic language model API."""
    url = "https://api.anthropic.com/v1/complete"
    headers = {
        "x-api-key": anthropic_key,
        "content-type": "application/json"
    }
    data = {
        "prompt": prompt,
        "model": "claude-v1.3",
        "max_tokens_to_sample": 1000,
        "temperature": 0.2,
        "stop_sequences": ["\n\nHuman: "]
    }
    response = requests.post(url, headers=headers, json=data)
    response.raise_for_status()
    return response.json()



Sample = collections.namedtuple('Sample', ['prompt', 'input', 'context', 'output'])


def try_candidate(prompt, inp):
    text, context = prompt.generator(inp)
    result = submit_prompt(text)
    return Sample(prompt, inp, context, result)


async def run_parallel_candidates(prompts):
    with ThreadPoolExecutor(max_workers=4) as executor:
        loop = asyncio.get_event_loop()
        return [loop.run_in_executor(executor, try_candidate, *prompt) for prompt in prompts]


# Evaluates whether A or B is better. Returns 1 if A is better, -1 if B is better.
def evaluate_candidates(conditions, goal, context, a, b):
    text = f'''I gave two people these high level instructions: "{goal}". I want you to evaluate who performed the task better. I gave them this context:

> ```
+ '\n> '.join(context.splitlines())
```

Person A wrote:

{a}

Person B wrote:

{b}

The goal was {goal}. {conditions}

Who performed the task better? Say A or B.

Assistant: Person '''
    result = submit_prompt(text)
    completion: str = result['completion'].lstrip()
    person_a = completion.find('A ')
    person_b = completion.find('B ')
    if person_a == -1 and person_b == -1:
        score = 0
    elif person_a == -1:
        score = -1
    elif person_b == -1:
        score = 1
    else:
        score = (person_a < person_b and 1) or -1
    #print(completion)
    #print(score)
    #print('-' * 10)
    return score


class WinLossTable(object):
    def __init__(self, n):
        self.n = n
        self.scores = [0 for _ in range(n*n)]

    def add(self, a, b, value):
        self.scores[(a * self.n) + b] += value
        self.scores[(b * self.n) + a] -= value

    def show(self):
        try:
            n = 1 + math.ceil(math.log10(max(map(abs, self.scores))))
        except ValueError:
            n = 2
        print('   |', end='')
        for i in range(self.n):
            print(f' {i:-{n}d}', end='')
        print()
        print('-' * ((self.n + 1) * (n + 1) + 2))
        for i in range(self.n):
            print(f'{i:-3d}|', end='')
            for j in range(self.n):
                print(f' {self.scores[(i * self.n) + j]:-{n}d}', end='')
            print(f'| {sum(self.scores[i*self.n:(i+1)*self.n]):-{n}d}')


def evaluate_and_score_candidates(table: WinLossTable, sample_a: Sample, sample_b: Sample):
    assert sample_a.input is sample_b.input
    assert sample_a.context == sample_b.context, 'could vary contexts but adjust evaluate_candidates to handle that fact'
    conditions = sample_a.input[0]
    goal = sample_a.input[1]
    score = evaluate_candidates(conditions, goal, sample_a.context, sample_a.output['completion'], sample_b.output['completion'])
    table.add(sample_a.prompt.id, sample_b.prompt.id, score)
    print(sample_a.prompt.id, sample_b.prompt.id, score)
    table.show()
    return score


async def run_parallel_evaluations(pairs):
    with ThreadPoolExecutor(max_workers=4) as executor:
        loop = asyncio.get_event_loop()
        return [loop.run_in_executor(executor, evaluate_and_score_candidates, *work) for work in pairs]


def main():
    print('Discovering prompts')
    schema = Schema()
    schema.discover_prompts(generate_prompt)
    for prompt in schema.prompts:
        print(f'{prompt.id}: {prompt.choices}')

    table = WinLossTable(len(schema.prompts))

    print('Generating candidate output')
    loop = asyncio.get_event_loop()
    corpus = list(map(lambda x: x.result(), loop.run_until_complete(run_parallel_candidates([(prompt, inp) for prompt in schema.prompts for inp in inputs]))))

    # Collect samples by input.
    samples_by_input = collections.defaultdict(dict)
    for sample in corpus:
        samples_by_input[sample.input][sample.prompt.id] = sample

    print('Comparing candidates')
    pairs = []
    for samples in samples_by_input.values():
        pairs.extend([(sA, sB) for sA in samples.values() for sB in samples.values() if sA is not sB])
    for result in loop.run_until_complete(run_parallel_evaluations([(table, *pair) for pair in pairs])):
        result.result()

    for prompt in schema.prompts:
        print(f'{prompt.id}: {prompt.choices}')
    table.show()

if __name__ == "__main__":
    main()