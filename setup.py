from setuptools import setup

setup(
    name='smashwalls',
    version='0.1',
    py_modules=['smash_walls'],
    install_requires=[
        'click',
        'cssselect',
        'lxml',
        'requests',
    ],
    entry_points='''
        [console_scripts]
        smashwalls=smash_walls:cli
    ''',
)
