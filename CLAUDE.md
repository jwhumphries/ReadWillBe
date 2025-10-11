This is a go based project. Create a GOTH stack application to help users track progress through a "daily reading plan". Some general design contraints are as follows:

# Function and Design
- Users should have the ability to register and create a username+password.
- Users should have the ability to "upload" a "daily reading plan" in the form of a CSV file.
- Users should have the ability to view any of their configured reading plan as a whole. The default view should just show users the current day.
- Users should have the ability to rename current plans, and delete current and old plans.
- Users should have the ability to configure push notificaitons that alert them of their reading plans for the day.
- Users should have the ability to mark a day's reading as complete.

## Webpage Views
- A login screen with standard "sign in" and "sign up" flows.
- A dashboard that shows the user any current readings that apply to today's date. Readings that are "out of date" (or not completed within their specified time period), should be displayed as well with some indication that the reading is overdue.
- A "history" view that shows previously completed readings.
- A "plans" view that allows users to rename plans, edit readings for future dates on a plan, or delete a plan entirely. There should also be a "create plan" view that allows users to upload a CSV file.
- An "account" view that allows users to manage push alerts, password, etc.

## Plans
- Plans should include a title and a list of "readings"
- "Readings" should include a date and a description of what to read.
  - The date can either be a month, a week, or a specific day.

# Tech Considerations
- Use @/Users/john/Documents/Code/lib/fanks as an example. Follow the patterns used in fanks.
- Use Templ, TailwindCSS, and HTMX for the webpages.
- Use the some of the styling like the font and colors from @/Users/john/Documents/Code/repos/jwh, but don't reference the jwh project for code examples.
- Use airflow for reloading.
- Prefer to build and run using docker and airflow. Use a Makefile to create useful commands.
- The database schema for a user's readings should look something like:
```
plan_name
date
reading
status
```
This should allow for the editing of readings by plan, by date, and for acting upon readings by their completion status. Keep in mind that date can be a month, week, or specific day.
- Plans uploaded via CSV file should be injested into memory, validated for proper format, then added to the database.
- Don't write comments for your code unless absolutely necesarry
- Write tests for your code
