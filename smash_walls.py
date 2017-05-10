#!/usr/bin/env python

import click
import datetime
import os
import re
import requests
from lxml import html


BASE_URL = 'https://www.smashingmagazine.com/tag/wallpapers/page/{}/'
PAGE_ANCHOR = '{month}-{year}'
MAX_PAGE = 12  # Fairly choosed by random
CALENDAR_DIR = '/cal/'
NO_CALENDAR_DIR = '/nocal/'
DEFAULT_DOWNLOAD_DIRECTORY = os.path.join(os.getenv('HOME', '.'), 'Pictures', 'Smashing-Wallpapers')
EXT_REGEXP = re.compile('\.(jpg|jpeg|png|gif)$')


def find_page_url(month, year):
    date = datetime.date(year=year, month=month, day=1)
    month_name = date.strftime("%B").lower()
    title = PAGE_ANCHOR.format(month=month_name, year=year)
    for page in range(MAX_PAGE):
        url = BASE_URL.format(page)
        response = requests.get(url)
        doc = html.fromstring(response.text)

        for a in doc.cssselect('article > h2 > a'):
            if title in a.attrib.get('href'):
                return a.attrib['href']


def download_wallpapers(month, year, download_folder, resolution, nocal):
    page_url = find_page_url(month, year)
    if not page_url:
        message = 'No page found for this month & year! :('
        click.echo(message, err=True)
        return 1

    click.echo('Starting download to {}'.format(download_folder))
    directory_check = NO_CALENDAR_DIR if nocal else CALENDAR_DIR

    response = requests.get(page_url)
    doc = html.fromstring(response.text)
    for anchor in doc.cssselect('li > a'):
        href = anchor.attrib['href']
        if not EXT_REGEXP.search(href):
            continue

        if directory_check not in href:
            continue

        if resolution not in href:
            continue

        filename = os.path.basename(href)
        download_file = os.path.join(download_folder, filename)
        with open(download_file, 'wb') as f:
            data = requests.get(href).content
            f.write(data)
            message = 'Successfully downloaded {}'.format(filename)
            click.echo(message)


@click.command()
@click.option('--month', help='Chosen month, defaults to current.')
@click.option('--year', help='Chosen year, defaults to current.')
@click.option('--dest', help='Custom download folder.')
@click.option('--res', default='1920x1080',
              help='Wallpaper resolution (default: 1920x1080).')
@click.option('--nocal', is_flag=True, help='Wallpaper without calendar.')
def cli(month, year, dest, res, nocal):
    """
    Simple program for downloading smashing magazine wallpapers
    to ~/Pictures/Smashing-Wallpapers/<year>/<month>
    """
    today = datetime.date.today()
    month = month or today.month
    year = year or today.year
    dest = dest or DEFAULT_DOWNLOAD_DIRECTORY

    month_dir = '{:02d}'.format(month)
    download_folder = os.path.join(dest, str(year), month_dir)

    try:
        os.makedirs(download_folder)
    except FileExistsError:
        # TODO: think if need to ask user
        pass

    return download_wallpapers(month, year, download_folder, res, nocal)


if __name__ == '__main__':
    cli()
