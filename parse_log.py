import re
import sys

pattern = re.compile(r'(?P<ts>\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}) (store|mqtt).*node=(?P<node>\S+) .*value=(?P<value>\S+)')

for line in sys.stdin:
    m = pattern.search(line)
    if m:
        ts = m.group('ts')
        node = m.group('node')
        value = m.group('value')
        print(f"{ts}\t{node}\t{value}")
